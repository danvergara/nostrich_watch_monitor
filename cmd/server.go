/*
Copyright © 2025 Daniel Vergara daniel.omar.vergara@gmail.com
*/
package cmd

import (
	"context"
	"fmt"
	"io/fs"
	"log/slog"
	"os"

	"github.com/spf13/cobra"

	"github.com/danvergara/nostrich_watch_monitor/internal/config"
	"github.com/danvergara/nostrich_watch_monitor/internal/handlers"
	"github.com/danvergara/nostrich_watch_monitor/internal/server"
	"github.com/danvergara/nostrich_watch_monitor/pkg/database"
	"github.com/danvergara/nostrich_watch_monitor/pkg/repository/postgres"
	"github.com/danvergara/nostrich_watch_monitor/pkg/services"
	"github.com/danvergara/nostrich_watch_monitor/web"
)

var (
	port string
)

//go:generate tailwindcss -i ./web/static/css/input.css -o ./web/static/css/styles.css --minify

// serverCmd represents the server command
var serverCmd = &cobra.Command{
	Use:   "server",
	Short: "HTTP server for the dashboard",
	RunE: func(cmd *cobra.Command, args []string) error {
		if port == "" {
			port = "8000"
		}

		jsonHandler := slog.NewJSONHandler(os.Stderr, nil)

		logger := slog.New(jsonHandler)

		staticFs, err := fs.Sub(web.StaticFiles, "static")
		if err != nil {
			return err
		}

		cfg := config.Config{
			Port:   port,
			Logger: logger,
		}

		dbConfig := database.Config{
			Host:     dbHost,
			Port:     dbPort,
			User:     dbUser,
			Password: dbPass,
			DBName:   dbName,
		}

		db, err := database.NewPostgresDB(dbConfig)
		if err != nil {
			return err
		}

		relayRepository := postgres.NewRelayRepository(db)
		relayService := services.NewRelayService(relayRepository, logger)
		relayHandler := handlers.NewRelaysHandler(relayService)

		logger.Info(fmt.Sprintf("Server listening on port %s", port))

		ctx := context.Background()
		if err := server.Run(ctx, &cfg, staticFs, *relayHandler); err != nil {
			logger.Error(fmt.Sprintf("Error running the server: %s", err))
			os.Exit(1)
		}

		return nil
	},
}

func init() {
	rootCmd.AddCommand(serverCmd)
	port = os.Getenv("DASHBOARD_SERVER_PORT")
}
