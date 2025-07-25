/*
Copyright © 2025 Daniel Vergara daniel.omar.vergara@gmail.com
*/
package cmd

import (
	"context"
	"fmt"
	"log"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/go-co-op/gocron/v2"
	"github.com/hibiken/asynq"
	"github.com/spf13/cobra"

	"github.com/danvergara/nostrich_watch_monitor/pkg/database"
	"github.com/danvergara/nostrich_watch_monitor/pkg/repository/postgres"
	"github.com/danvergara/nostrich_watch_monitor/pkg/task"
)

// schedulerCmd represents the scheduler command
var schedulerCmd = &cobra.Command{
	Use:   "scheduler",
	Short: "scheduler command automatically enqueues tasks to monitor relays.",
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx := context.Background()

		opts := &slog.HandlerOptions{
			Level:     slog.LevelDebug, // Set minimum log level
			AddSource: true,            // Include source code location
		}

		logger := slog.New(slog.NewJSONHandler(os.Stdout, opts))

		client := asynq.NewClient(asynq.RedisClientOpt{Addr: redisHost})

		dbConfig := database.Config{
			Host:     dbHost,
			Port:     dbPort,
			User:     dbUser,
			Password: dbPass,
			DBName:   dbName,
		}

		// Create a PostgreSQL database pool of connections given config data.
		db, err := database.NewPostgresDB(dbConfig)
		if err != nil {
			return err
		}
		defer db.Close()

		// Use the pool of connections to create a database client, based on the Repository pattern.
		relayRepo := postgres.NewRelayRepository(db)

		// Retrieve a list of relays to perfom health checks on.
		relays, err := relayRepo.List(ctx)
		if err != nil {
			return err
		}

		// Create a context for graceful shutdown.
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		// Create a Cron Job scheduler.
		s, err := gocron.NewScheduler()
		if err != nil {
			return err
		}

		// Ensure shutdown when done
		defer func() {
			if err := s.Shutdown(); err != nil {
				logger.Error(fmt.Sprintf("error shutting down scheduler: %v", err))
			}
		}()

		// Create a job for each relay.
		jobs := make([]gocron.Job, 0, len(relays))

		for _, r := range relays {
			// Create a job for this specific relay
			job, err := s.NewJob(
				gocron.DurationJob(1*time.Hour),
				gocron.NewTask(func(relayURL string) error {
					// Create a asynq task passing the type and the payload of the task.
					relayTask, err := task.NewRelayHealthCheckTask(relayURL)
					if err != nil {
						logger.Error(err.Error())
						return err
					}

					// Process the task immediately.
					info, err := client.Enqueue(relayTask)
					if err != nil {
						logger.Error(fmt.Sprintf("error processing a task: %s", err))
						return err
					}

					logger.Info(fmt.Sprintf("[*] Successfully enqueued the task: %+v", info))

					return nil

				}, r.URL),
				gocron.WithContext(ctx),
				gocron.WithName(fmt.Sprintf("health check on relay: %s", r.URL)),
				gocron.WithTags("health-check", "monitoring", r.URL),
			)
			if err != nil {
				logger.Error(fmt.Sprintf("error scheduling job: %v", err))
				continue
			}

			jobs = append(jobs, job)
		}

		jobAnnouncement, err := s.NewJob(
			gocron.DurationJob(7*24*time.Hour),
			gocron.NewTask(func(frequency string) error {
				// Create a asynq task passing the type and the payload of the task.
				relayTask, err := task.NewTaskMonitorAnnouncement(frequency)
				if err != nil {
					logger.Error(err.Error())
					return err
				}

				// Process the task immediately.
				info, err := client.Enqueue(relayTask)
				if err != nil {
					logger.Error(fmt.Sprintf("error processing a task: %s", err))
					return err
				}

				logger.Info(fmt.Sprintf("[*] Successfully enqueued the task: %+v", info))

				return nil

			}, "604800"),
			gocron.WithContext(ctx),
			gocron.WithName("Monitor Announcement"),
			gocron.WithTags("monitoring", "announcement"),
		)

		jobs = append(jobs, jobAnnouncement)
		// Start the scheduler.
		s.Start()
		logger.Info("scheduler started. Task will run every 15 minutes.")

		// Show next run times.
		logger.Info("next run times:")
		for _, job := range jobs {
			if nextRun, err := job.NextRun(); err == nil {
				logger.Info(
					fmt.Sprintf("%s: %s", job.Name(), nextRun.Format("2006-01-02 15:04:05")),
				)
			}
		}

		// Set up graceful shutdown.
		c := make(chan os.Signal, 1)
		signal.Notify(c, os.Interrupt, syscall.SIGTERM)

		log.Println("Press Ctrl+C to stop...")

		<-c

		logger.Info("shutting down health check scheduler...")

		return nil
	},
}

func init() {
	rootCmd.AddCommand(schedulerCmd)
}
