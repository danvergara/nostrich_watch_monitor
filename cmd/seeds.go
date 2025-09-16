/*
Copyright © 2025 Daniel Vergara daniel.omar.vergara@gmail.com
*/
package cmd

import (
	"context"
	"fmt"
	"log"

	"github.com/jmoiron/sqlx"
	"github.com/lib/pq"
	_ "github.com/lib/pq"
	"github.com/nbd-wtf/go-nostr/nip11"
	"github.com/spf13/cobra"

	"github.com/danvergara/nostrich_watch_monitor/pkg/domain"
	"github.com/danvergara/nostrich_watch_monitor/pkg/repository/postgres"
)

func seedRelays(db *sqlx.DB) error {
	relayURLs := []string{
		"wss://relay.damus.io",
		"wss://relay.nostr.band",
		"wss://nostr.land",
		"wss://nostr.mom",
		"wss://nos.lol",
	}

	ctx := context.Background()

	for _, url := range relayURLs {
		info, err := nip11.Fetch(ctx, url)
		if err != nil {
			log.Printf("failed to fetch nip 11 info for relay %s: %v", url, err)
			continue
		}

		// If NIP-11 was successful, update relay metadata.
		supportedNIPsSlice, err := convertAnyToInt(info.SupportedNIPs)
		if err != nil {
			log.Printf(
				"failed to handle supported NIPs slice from nip 11 info for relay %s: %v",
				url,
				err,
			)
			continue
		}

		// Convert []int to pq.Int64Array
		var supportedNIPs pq.Int64Array
		for _, nip := range supportedNIPsSlice {
			supportedNIPs = append(supportedNIPs, int64(nip))
		}

		relayInfo := domain.Relay{
			URL:            url,
			Name:           &info.Name,
			Description:    &info.Description,
			PubKey:         &info.PubKey,
			Contact:        &info.Contact,
			SupportedNIPs:  supportedNIPs,
			Software:       &info.Software,
			Version:        &info.Version,
			Icon:           &info.Icon,
			Banner:         &info.Banner,
			PostingPolicy:  &info.PostingPolicy,
			Tags:           pq.StringArray(info.Tags),
			LanguageTags:   pq.StringArray(info.LanguageTags),
			RelayCountries: pq.StringArray(info.RelayCountries),
		}

		relayRepo := postgres.NewRelayRepository(db)

		if err := relayRepo.Create(ctx, relayInfo); err != nil {
			log.Printf("failed to insert %s: %v", url, err)
		}
	}

	return nil
}

// seedsCmd represents the seeds command
var seedsCmd = &cobra.Command{
	Use:   "seeds",
	Short: "Seeds the database with relay data",
	RunE: func(cmd *cobra.Command, args []string) error {
		db, err := sqlx.Open("postgres",
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

// convertAnyToInt utility function to convert an slice of any to an slice of integers.
// Moslty used for supported NIPS from the NIP 11 response.
func convertAnyToInt(input []any) ([]int, error) {
	var result []int

	for i, value := range input {
		switch v := value.(type) {
		case int:
			result = append(result, v)
		case float64:
			result = append(result, int(v))
		default:
			return nil, fmt.Errorf("element at index %d not supported - real type is %T", i, value)
		}
	}

	return result, nil
}
