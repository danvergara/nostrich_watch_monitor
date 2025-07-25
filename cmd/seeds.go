/*
Copyright © 2025 Daniel Vergara daniel.omar.vergara@gmail.com
*/
package cmd

import (
	"database/sql"
	"fmt"
	"log"

	_ "github.com/lib/pq"
	"github.com/spf13/cobra"
)

func seedRelays(db *sql.DB) error {
	relayURLs := []string{
		"wss://relay.damus.io",
		"wss://relay.nostr.band",
		"wss://nostr.land",
		"wss://nostr.mom",
		"wss://nos.lol",
	}

	stmt, err := db.Prepare(`
        INSERT INTO relays (url, created_at, updated_at) 
        VALUES ($1, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP)
        ON CONFLICT (url) DO NOTHING
    `)
	if err != nil {
		return err
	}

	defer func() {
		_ = stmt.Close()
	}()

	for _, url := range relayURLs {
		_, err := stmt.Exec(url)
		if err != nil {
			log.Printf("failed to insert %s: %v", url, err)
		} else {
			log.Printf("seeded relay: %s", url)
		}
	}

	return nil
}

// seedsCmd represents the seeds command
var seedsCmd = &cobra.Command{
	Use:   "seeds",
	Short: "Seeds the database with relay data",
	RunE: func(cmd *cobra.Command, args []string) error {
		db, err := sql.Open("postgres",
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
		defer func() {
			_ = db.Close()
		}()

		if err := seedRelays(db); err != nil {
			return err
		}

		return nil
	},
}

func init() {
	rootCmd.AddCommand(seedsCmd)
}
