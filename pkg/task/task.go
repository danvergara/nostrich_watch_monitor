package task

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"time"

	"github.com/hibiken/asynq"
	"github.com/jmoiron/sqlx"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"golang.org/x/sys/unix"

	"github.com/danvergara/nostrich_watch_monitor/pkg/healthcheck"
)

const (
	TypeHealthCheck         = "relay:healthcheck"
	TypeMonitorAnnouncement = "relay:announcement"
)

// Metric variables.
var (
	processedCounter = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "processed_tasks_total",
			Help: "Total number of processed tasks",
		},
		[]string{"task_type"},
	)

	failedCounter = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "failed_tasks_total",
			Help: "Total number of failed tasks processed",
		},
		[]string{"task_type"},
	)

	inProgressGauge = promauto.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "in_progress_tasks",
			Help: "Current number of tasks being processed",
		},
		[]string{"task_type"},
	)
)

type TasKHandler struct {
	db           *sqlx.DB
	timeout      time.Duration
	privateKey   string // For signing test events
	logger       *slog.Logger
	redisHost    string
	monitorRelay string
}

func NewTaskHandler(
	db *sqlx.DB,
	timeout time.Duration,
	privateKey string,
	logger *slog.Logger,
	redisHost string,
	monitorRelayURL string,
) *TasKHandler {
	return &TasKHandler{
		db:           db,
		timeout:      timeout,
		privateKey:   privateKey,
		logger:       logger,
		redisHost:    redisHost,
		monitorRelay: monitorRelayURL,
	}
}

func (th *TasKHandler) Run() error {
	httpServeMux := http.NewServeMux()
	httpServeMux.Handle("/metrics", promhttp.Handler())
	metricsSrv := &http.Server{
		Addr:    ":2112",
		Handler: httpServeMux,
	}
	done := make(chan struct{})

	// Start the metrics server.
	go func() {
		if err := metricsSrv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			th.logger.Error("metrics server errored", slog.Any("error", err.Error()))
		}

		close(done)
	}()

	srv := asynq.NewServer(
		asynq.RedisClientOpt{Addr: th.redisHost},
		asynq.Config{Concurrency: 10},
	)

	mux := asynq.NewServeMux()
	mux.Use(metricsMiddleware)
	mux.HandleFunc(TypeHealthCheck, th.HandleRelayHealthCheckTask)
	mux.HandleFunc(TypeMonitorAnnouncement, th.HandleMonitorAnnouncementTask)

	if err := srv.Start(mux); err != nil {
		th.logger.Error("Failed to start worker server", slog.Any("error", err.Error()))
		return err
	}

	// Wait for termination signals.
	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, unix.SIGTERM, unix.SIGINT)

	select {
	case sig := <-sigs:
		th.logger.Info("received signal", slog.Any("signal", sig))
	case <-done:
		th.logger.Info(
			"metrics server shut down unexpectedly",
		)
	}

	srv.Stop()
	th.logger.Info("worker server stopped")

	return nil
}

// Payload for any task related to health checks on Nostr relays.
type RelayHealthCheckTaskPayload struct {
	// URL of the relay
	RelayURL string
}

type RelayMonitorAnnouncementTaskPayload struct {
	Frequency string
}

func (th *TasKHandler) HandleRelayHealthCheckTask(ctx context.Context, t *asynq.Task) error {
	var r RelayHealthCheckTaskPayload

	if err := json.Unmarshal(t.Payload(), &r); err != nil {
		return err
	}

	rc := healthcheck.NewRelayChecker(
		healthcheck.WithDB(th.db),
		healthcheck.WithTimeout(th.timeout),
		healthcheck.WithPrivateKey(th.privateKey),
		healthcheck.WithLogger(th.logger),
		healthcheck.WithMonitorRelay(th.monitorRelay),
	)
	if err := rc.CheckRelay(ctx, r.RelayURL); err != nil {
		return err
	}

	th.logger.Info("[*] health checking", slog.String("nostr_relay", r.RelayURL))

	return nil
}

func (th *TasKHandler) HandleMonitorAnnouncementTask(ctx context.Context, t *asynq.Task) error {
	var r RelayMonitorAnnouncementTaskPayload

	if err := json.Unmarshal(t.Payload(), &r); err != nil {
		return err
	}

	rc := healthcheck.NewRelayChecker(
		healthcheck.WithTimeout(th.timeout),
		healthcheck.WithPrivateKey(th.privateKey),
		healthcheck.WithLogger(th.logger),
		healthcheck.WithMonitorRelay(th.monitorRelay),
	)

	// convert the timeout to seconds with no decimals and then to string to be passed to the creationg of the 10166 event.
	timeoutInSeconds := fmt.Sprintf("%d", int64(th.timeout.Seconds()))

	if err := rc.Publish10166Event(ctx, r.Frequency, timeoutInSeconds); err != nil {
		return err
	}

	return nil
}

func metricsMiddleware(next asynq.Handler) asynq.Handler {
	return asynq.HandlerFunc(func(ctx context.Context, t *asynq.Task) error {
		inProgressGauge.WithLabelValues(t.Type()).Inc()
		err := next.ProcessTask(ctx, t)
		inProgressGauge.WithLabelValues(t.Type()).Dec()
		if err != nil {
			failedCounter.WithLabelValues(t.Type()).Inc()
		}
		processedCounter.WithLabelValues(t.Type()).Inc()
		return err
	})
}

func NewRelayHealthCheckTask(relayURL string) (*asynq.Task, error) {
	payload, err := json.Marshal(RelayHealthCheckTaskPayload{RelayURL: relayURL})
	if err != nil {
		return nil, err
	}

	return asynq.NewTask(TypeHealthCheck, payload), nil
}

func NewTaskMonitorAnnouncement(frequency string) (*asynq.Task, error) {
	payload, err := json.Marshal(RelayMonitorAnnouncementTaskPayload{Frequency: frequency})
	if err != nil {
		return nil, err
	}

	return asynq.NewTask(TypeMonitorAnnouncement, payload), nil
}
