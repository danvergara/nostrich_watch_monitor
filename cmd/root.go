/*
Copyright © 2025 Daniel Vergara daniel.omar.vergara@gmail.com
*/
package cmd

import (
	"os"

	"github.com/spf13/cobra"
)

var (
	dbHost    string
	dbPort    string
	dbUser    string
	dbPass    string
	dbName    string
	redisHost string
)

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "monitor",
	Short: "Nostr Relay checker",
	Long:  `A NIP 66 compatible Nostr Relay health checker that publishes relays statuses as 30166 events`,
	RunE: func(cmd *cobra.Command, args []string) error {
		return nil
	},
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	cobra.CheckErr(rootCmd.Execute())
}

func init() {
	// Initialize the redisHost variable based on the value of the NOSTRICH_WATCH_REDIS_HOST environment variable.
	redisHost = os.Getenv("NOSTRICH_WATCH_REDIS_HOST")
	dbHost = os.Getenv("NOSTRICH_WATCH_DB_HOST")
	dbPort = os.Getenv("NOSTRICH_WATCH_DB_PORT")
	dbUser = os.Getenv("NOSTRICH_WATCH_DB_USER")
	dbPass = os.Getenv("NOSTRICH_WATCH_DB_PASSWORD")
	dbName = os.Getenv("NOSTRICH_WATCH_DB_NAME")
}
