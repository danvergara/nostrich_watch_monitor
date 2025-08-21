package postgres

import (
	"context"
	"fmt"

	"github.com/Masterminds/squirrel"
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
func (r *relayRepository) List(
	ctx context.Context,
	opts *repository.ListOption,
) ([]domain.Relay, error) {
	var relays = []domain.Relay{}

	// Select automatically scans the query results into a slice of structs.
	query := squirrel.Select(
		"url",
		"name",
		"description",
		"pubkey",
		"contact",
		"supported_nips",
		"software",
		"version",
		"icon",
		"banner",
		"privacy_policy",
		"terms_of_service",
		"relay_countries",
		"language_tags",
		"tags",
		"posting_policy",
	).
		From("relays").
		PlaceholderFormat(squirrel.Dollar)

	if opts != nil {
		if opts.Limit != nil {
			query = query.Limit(uint64(*opts.Limit))
		}
		if opts.Offset != nil {
			query = query.Offset(uint64(*opts.Offset))
		}
	}

	sql, args, err := query.ToSql()
	if err != nil {
		return nil, fmt.Errorf("failed to build query: %w", err)
	}

	if err := r.db.SelectContext(ctx, &relays, sql, args...); err != nil {
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
