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

	psql := sq.StatementBuilder.PlaceholderFormat(sq.Dollar)

	// Build the subquery for latest health checks.
	subquery := psql.Select(
		"DISTINCT ON (relay_url) relay_url",
		"created_at",
		"websocket_success",
		"nip11_success",
		"rtt_open",
	).From("health_checks").
		OrderBy("relay_url", "created_at DESC")

	subquerySQL, subqueryArgs, err := subquery.ToSql()
	if err != nil {
		return nil, fmt.Errorf("failed to build subquery: %w", err)
	}

	// Select automatically scans the query results into a slice of structs.
	query := sq.Select(
		"r.*",
		`h.created_at AS "health_check.created_at"`,
		`h.websocket_success AS "health_check.websocket_success"`,
		`h.nip11_success AS "health_check.nip11_success"`,
		`h.rtt_open AS "health_check.rtt_open"`,
	).
		From("relays AS r").
		LeftJoin(fmt.Sprintf("(%s) h ON r.url = h.relay_url", subquerySQL)).
		OrderBy("r.url")

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

	sql, mainArgs, err := query.ToSql()
	if err != nil {
		return nil, fmt.Errorf("failed to build main query: %w", err)
	}

	// Combine arguments.
	args := append(subqueryArgs, mainArgs...)

	if err := r.db.SelectContext(ctx, &relays, sql, args...); err != nil {
		return nil, fmt.Errorf("failed to get relays: %w", err)
	}

	return relays, nil
}

func (r *relayRepository) FindByURL(ctx context.Context, url string) (domain.Relay, error) {
	relay := domain.Relay{}

	psql := sq.StatementBuilder.PlaceholderFormat(sq.Dollar)

	query := psql.Select(
		"r.*",
		`h.created_at AS "health_check.created_at"`,
		`h.websocket_success AS "health_check.websocket_success"`,
		`h.nip11_success AS "health_check.nip11_success"`,
		`h.rtt_open AS "health_check.rtt_open"`,
	).From("relay AS r").
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
