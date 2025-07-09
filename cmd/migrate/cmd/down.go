/*
Copyright © 2025 Daniel Vergara daniel.omar.vergara@gmail.com
*/
package cmd

import (
	"fmt"
	"os"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/spf13/cobra"
)

// downCmd represents the down command
var downCmd = &cobra.Command{
	Use:   "down",
	Short: "Rollback the database migrations",
	RunE: func(cmd *cobra.Command, args []string) error {
		m, err := migrate.New(
			"file://db/migrations",
			fmt.Sprintf(
				"postgres://%s:%s@%s:%s/%s?sslmode=disable",
				dbUser,
				dbPass,
				dbHost,
				dbPort,
				dbName,
			),
		)
		if err != nil {
			return err
		}

		if err := m.Down(); err != nil {
			return err
		}

		return nil

	},
}

func init() {
	dbHost = os.Getenv("NOSTRICH_WATCH_DB_HOST")
	dbPort = os.Getenv("NOSTRICH_WATCH_DB_PORT")
	dbUser = os.Getenv("NOSTRICH_WATCH_DB_USER")
	dbPass = os.Getenv("NOSTRICH_WATCH_DB_PASSWORD")
	dbName = os.Getenv("NOSTRICH_WATCH_DB_NAME")

	migrateCmd.AddCommand(downCmd)
}
