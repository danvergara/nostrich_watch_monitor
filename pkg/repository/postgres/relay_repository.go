package postgres

import (
	"context"
	"fmt"

	"github.com/jmoiron/sqlx"

	"github.com/danvergara/nostrich_watch_monitor/pkg/domain"
	"github.com/danvergara/nostrich_watch_monitor/pkg/repository"
)

type relayRepository struct {
	db *sqlx.DB
}

func NewRelayRepository(db *sqlx.DB) repository.RelayRepository {
	return &relayRepository{db: db}
}

// List returns a list of relays of interest from the databased to be monitored.
func (r *relayRepository) List(ctx context.Context) ([]domain.Relay, error) {
	var relays = []domain.Relay{}

	// Select automatically scans the query results into a slice of structs.
	err := r.db.SelectContext(ctx, &relays, "SELECT url FROM relays")
	if err != nil {
		return nil, err
	}

	return relays, nil
}

func (r *relayRepository) Update(ctx context.Context, relayInfo domain.Relay) error {
	query := `
        UPDATE relays SET 
            name = :name,
            description = :description,
            pubkey = :pubkey,
            contact = :contact,
            supported_nips = :supported_nips,
            software = :software,
            version = :version,
            icon = :icon,
            banner = :banner,
            privacy_policy = :privacy_policy,
            terms_of_service = :terms_of_service,
            posting_policy = :posting_policy,
            updated_at = CURRENT_TIMESTAMP
        WHERE url = :url
    `

	if _, err := r.db.NamedExec(query, relayInfo); err != nil {
		return fmt.Errorf("failed to update relay info: %w", err)
	}

	return nil
}

func (r *relayRepository) SaveHealthCheck(ctx context.Context, status domain.HealthCheck) error {
	query := `
        INSERT INTO health_checks (
            relay_url, created_at, websocket_success, websocket_error,
            nip11_success, nip11_error, rtt_open, rtt_read, rtt_write, rtt_nip11
        ) VALUES (
					:relay_url, :created_at, :websocket_success, :websocket_error,
        	:nip11_success, :nip11_error, :rtt_open, :rtt_read, :rtt_write, :rtt_nip11
				)
    `
	_, err := r.db.NamedExec(query, status)
	if err != nil {
		return fmt.Errorf("failed to save health check: %w", err)
	}

	return nil
}
