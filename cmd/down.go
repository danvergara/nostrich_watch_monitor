/*
Copyright © 2025 Daniel Vergara daniel.omar.vergara@gmail.com
*/
package cmd

import (
	"fmt"
	"log"

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

		log.Println("database migrations were reversed succesfully")

		return nil
	},
}

func init() {
	migrateCmd.AddCommand(downCmd)
}
