package task

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"time"

	"github.com/hibiken/asynq"
	"github.com/jmoiron/sqlx"

	"github.com/danvergara/nostrich_watch_monitor/pkg/healthcheck"
)

const (
	TypeHealthCheck = "relay:healthcheck"
)

type TasKHandler struct {
	db         *sqlx.DB
	timeout    time.Duration
	privateKey string // For signing test events
	logger     *slog.Logger
	redisHost  string
}

func NewTaskHandler(
	db *sqlx.DB,
	timeout time.Duration,
	privateKey string,
	logger *slog.Logger,
	redisHost string,
) *TasKHandler {
	return &TasKHandler{
		db:         db,
		timeout:    timeout,
		privateKey: privateKey,
		logger:     logger,
		redisHost:  redisHost,
	}
}

func (th *TasKHandler) Run() error {
	srv := asynq.NewServer(
		asynq.RedisClientOpt{Addr: th.redisHost},
		asynq.Config{Concurrency: 10},
	)

	mux := asynq.NewServeMux()
	mux.HandleFunc(TypeHealthCheck, th.HandleRelayHealthCheckTask)

	if err := srv.Run(mux); err != nil {
		th.logger.Error(err.Error())
		return err
	}

	return nil
}

// Payload for any task related to health checks on Nostr relays.
type RelayHealthCheckTaskPayload struct {
	// URL of the relay
	RelayURL string
}

func (th *TasKHandler) HandleRelayHealthCheckTask(ctx context.Context, t *asynq.Task) error {
	var r RelayHealthCheckTaskPayload

	if err := json.Unmarshal(t.Payload(), &r); err != nil {
		return err
	}

	rc := healthcheck.NewRelayChecker(th.db, th.timeout, th.privateKey, th.logger)
	if err := rc.CheckRelay(ctx, r.RelayURL); err != nil {
		return err
	}

	th.logger.Info(fmt.Sprintf("[*] health checking %s Nostr relay", r.RelayURL))

	return nil
}

func NewRelayHealthCheckTask(relayURL string) (*asynq.Task, error) {
	payload, err := json.Marshal(RelayHealthCheckTaskPayload{RelayURL: relayURL})
	if err != nil {
		return nil, err
	}

	return asynq.NewTask(TypeHealthCheck, payload), nil
}
