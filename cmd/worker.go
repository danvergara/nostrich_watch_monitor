/*
Copyright © 2025 Daniel Vergara daniel.omar.vergara@gmail.com
*/
package cmd

import (
	"log/slog"
	"os"

	"github.com/hibiken/asynq"
	"github.com/spf13/cobra"

	"github.com/danvergara/nostrich_watch_monitor/pkg/task"
)

// workerCmd represents the worker command
var workerCmd = &cobra.Command{
	Use:   "worker",
	Short: "The worker command performs health checks on relays",
	RunE: func(cmd *cobra.Command, args []string) error {
		opts := &slog.HandlerOptions{
			Level:     slog.LevelDebug, // Set minimum log level
			AddSource: true,            // Include source code location
		}

		logger := slog.New(slog.NewJSONHandler(os.Stdout, opts))

		srv := asynq.NewServer(
			asynq.RedisClientOpt{Addr: redisHost},
			asynq.Config{Concurrency: 10},
		)

		mux := asynq.NewServeMux()
		mux.HandleFunc(task.TypeHealthCheck, task.HandleRelayHealthCheckTask)

		if err := srv.Run(mux); err != nil {
			logger.Error(err.Error())
			return err
		}

		return nil
	},
}

func init() {
	rootCmd.AddCommand(workerCmd)
}
