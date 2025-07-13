/*
Copyright © 2025 Daniel Vergara daniel.omar.vergara@gmail.com
*/
package cmd

import (
	"context"
	"os"

	"github.com/spf13/cobra"

	"github.com/danvergara/nostrich_watch_monitor/pkg/database"
	"github.com/danvergara/nostrich_watch_monitor/pkg/repository/postgres"
)

var (
	dbHost string
	dbPort string
	dbUser string
	dbPass string
	dbName string
)

// schedulerCmd represents the scheduler command
var schedulerCmd = &cobra.Command{
	Use:   "scheduler",
	Short: "scheduler command automatically enqueues tasks to monitor relays.",
	Long:  `.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx := context.Background()

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
		defer db.Close()

		relayRepo := postgres.NewRelayRepository(db)
		relayRepo.List(ctx)

		return nil
	},
}

func init() {
	dbHost = os.Getenv("NOSTRICH_WATCH_DB_HOST")
	dbPort = os.Getenv("NOSTRICH_WATCH_DB_PORT")
	dbUser = os.Getenv("NOSTRICH_WATCH_DB_USER")
	dbPass = os.Getenv("NOSTRICH_WATCH_DB_PASSWORD")
	dbName = os.Getenv("NOSTRICH_WATCH_DB_NAME")

	rootCmd.AddCommand(schedulerCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// schedulerCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// schedulerCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
