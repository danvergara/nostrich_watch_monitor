package postgres

import (
	"context"
	"fmt"

	sq "github.com/Masterminds/squirrel"
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
	var relays []domain.Relay

	tx, err := r.db.BeginTxx(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	if _, err := tx.ExecContext(ctx, "SELECT setseed(EXTRACT(DOY FROM current_date) / 366.0)"); err != nil {
		return nil, fmt.Errorf("failed to set seed: %w", err)
	}

	// Build the complete query with subquery inline
	query := sq.StatementBuilder.PlaceholderFormat(sq.Dollar).Select(
		"r.url",
		"r.name",
		"r.description",
		"r.pubkey",
		"r.contact",
		"r.icon",
		"r.banner",
		"r.privacy_policy",
		"r.terms_of_service",
		"r.software",
		"r.version",
		"r.supported_nips",
		"r.relay_countries",
		"r.language_tags",
		"r.tags",
		"r.posting_policy",
		"r.created_at",
		"r.updated_at",
		`h.created_at AS "health_checks.created_at"`,
		`h.websocket_success AS "health_checks.websocket_success"`,
		`h.nip11_success AS "health_checks.nip11_success"`,
		`h.rtt_open AS "health_checks.rtt_open"`,
		`h.rtt_nip11 AS "health_checks.rtt_nip11"`,
	).
		From("relays AS r").
		LeftJoin(`(
			SELECT DISTINCT ON (relay_url) relay_url, created_at, websocket_success, nip11_success, rtt_open, rtt_nip11
			FROM health_checks
			ORDER BY relay_url, created_at DESC
		) h ON r.url = h.relay_url`)

	if opts != nil {
		if opts.Limit != nil {
			query = query.Limit(uint64(*opts.Limit))
		}
		if opts.Offset != nil {
			query = query.Offset(uint64(*opts.Offset))
		}

		if len(opts.URLs) > 0 {
			query = query.Where(sq.Eq{"r.url": opts.URLs})
		}
	}

	query = query.OrderBy("RANDOM()")

	sql, args, err := query.ToSql()
	if err != nil {
		return nil, fmt.Errorf("failed to build query: %w", err)
	}

	if err := tx.SelectContext(ctx, &relays, sql, args...); err != nil {
		return nil, fmt.Errorf("failed to get relays: %w", err)
	}

	if err := tx.Commit(); err != nil {
		return nil, fmt.Errorf("failed to commit transaction: %w", err)
	}

	// Post-process to set HealthCheck to nil when there's no actual health check data
	for i := range relays {
		if relays[i].HealthCheck != nil && relays[i].HealthCheck.CreatedAt == nil {
			relays[i].HealthCheck = nil
		}
	}

	return relays, nil
}

func (r *relayRepository) FindByURL(ctx context.Context, url string) (domain.Relay, error) {
	relay := domain.Relay{}

	psql := sq.StatementBuilder.PlaceholderFormat(sq.Dollar)

	query := psql.Select(
		"r.url",
		"r.name",
		"r.description",
		"r.pubkey",
		"r.contact",
		"r.icon",
		"r.banner",
		"r.privacy_policy",
		"r.terms_of_service",
		"r.software",
		"r.version",
		"r.supported_nips",
		"r.relay_countries",
		"r.language_tags",
		"r.tags",
		"r.posting_policy",
		"r.created_at",
		"r.updated_at",
		`h.created_at AS "health_checks.created_at"`,
		`h.websocket_success AS "health_checks.websocket_success"`,
		`h.nip11_success AS "health_checks.nip11_success"`,
		`h.rtt_open AS "health_checks.rtt_open"`,
		`h.rtt_nip11 AS "health_checks.rtt_nip11"`,
	).From("relays AS r").
		LeftJoin("health_checks AS h ON r.url = h.relay_url").
		Where(sq.Eq{"r.url": url}).
		OrderBy("h.created_at DESC").
		Limit(1)

	sql, args, err := query.ToSql()
	if err != nil {
		return domain.Relay{}, fmt.Errorf("failed to build query: %w", err)
	}

	if err := r.db.GetContext(ctx, &relay, sql, args...); err != nil {
		return domain.Relay{}, fmt.Errorf("not found %w", err)
	}

	// Set HealthCheck to nil when there's no actual health check data
	if relay.HealthCheck != nil && relay.HealthCheck.CreatedAt == nil {
		relay.HealthCheck = nil
	}

	return relay, nil
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
        WHERE url = :url`

	if _, err := r.db.NamedExec(query, relayInfo); err != nil {
		return fmt.Errorf("failed to update relay info: %w", err)
	}

	return nil
}

func (r *relayRepository) Create(ctx context.Context, relayInfo domain.Relay) error {
	query := `
		INSERT INTO relays (
			url,
			name,
			description,
			pubkey,
			contact,
			supported_nips,
			software,
			version,
			icon,
			banner,
			privacy_policy,
			terms_of_service,
			posting_policy,
			updated_at
		)
		VALUES (
			:url,
			:name,
			:description,
			:pubkey,
			:contact,
			:supported_nips,
			:software,
			:version,
			:icon,
			:banner,
			:privacy_policy,
			:terms_of_service,
			:posting_policy,
			:updated_at
		)
		ON CONFLICT (url) DO NOTHING`

	if _, err := r.db.NamedExecContext(ctx, query, relayInfo); err != nil {
		return fmt.Errorf("failed to insert relay info: %w", err)
	}

	return nil
}

func (r *relayRepository) SaveHealthCheck(ctx context.Context, status domain.HealthCheck) error {
	query := `
        INSERT INTO health_checks (
          relay_url,
					created_at,
					websocket_success,
					websocket_error,
          nip11_success,
					nip11_error,
					rtt_open,
					rtt_read,
					rtt_write,
					rtt_nip11
        )
				VALUES (
					:relay_url,
					:created_at,
					:websocket_success,
					:websocket_error,
        	:nip11_success,
					:nip11_error,
					:rtt_open,
					:rtt_read,
					:rtt_write,
					:rtt_nip11
				)`

	_, err := r.db.NamedExec(query, status)
	if err != nil {
		return fmt.Errorf("failed to save health check: %w", err)
	}

	return nil
}
