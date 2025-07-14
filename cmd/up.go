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

// upCmd represents the up command
var upCmd = &cobra.Command{
	Use:   "up",
	Short: "Run the Nostrich Watch database migrations",
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

		if err := m.Up(); err != nil {
			return err
		}

		log.Println("migrations applied succesfully")

		return nil
	},
}

func init() {
	migrateCmd.AddCommand(upCmd)
}
