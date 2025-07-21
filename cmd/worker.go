/*
Copyright © 2025 Daniel Vergara daniel.omar.vergara@gmail.com
*/
package cmd

import (
	"log/slog"
	"os"
	"time"

	"github.com/spf13/cobra"

	"github.com/danvergara/nostrich_watch_monitor/pkg/database"
	"github.com/danvergara/nostrich_watch_monitor/pkg/task"
)

var monitorPrivateKey string

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

		th := task.NewTaskHandler(db, 10*time.Second, monitorPrivateKey, logger, redisHost)

		if err := th.Run(); err != nil {
			logger.Error(err.Error())
			return err
		}

		return nil
	},
}

func init() {
	monitorPrivateKey = os.Getenv("NOSTRICH_WATCH_MONITOR_PRIVATE_KEY")
	rootCmd.AddCommand(workerCmd)
}
