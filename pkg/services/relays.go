package services

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/danvergara/nostrich_watch_monitor/pkg/domain"
	"github.com/danvergara/nostrich_watch_monitor/pkg/repository"
)

type RelayService interface {
	GetRelayByURL(context.Context, string) (domain.Relay, error)
	GetRelays(context.Context, *RelayFilters) ([]domain.Relay, error)
}

type RelayFilters struct {
	Limit  *int
	Offset *int
	URLs   []string
}

type relayService struct {
	relayRepo repository.RelayRepository
	logger    *slog.Logger
}

func NewRelayService(
	relayRepo repository.RelayRepository,
	logger *slog.Logger,
) RelayService {
	return &relayService{
		relayRepo: relayRepo,
		logger:    logger,
	}
}

func (rs *relayService) GetRelayByURL(ctx context.Context, url string) (domain.Relay, error) {
	rs.logger.Info("Fetching relay", slog.String("url", url))

	r, err := rs.relayRepo.FindByURL(ctx, url)
	if err != nil {
		rs.logger.Error("Failed to find relay",
			slog.String("url", url),
			slog.String("error", err.Error()),
		)
		return domain.Relay{}, fmt.Errorf("could not find relay %s: %w", url, err)
	}

	rs.logger.Info("Successfully fetched relay",
		slog.String("url", url),
		slog.String("name", func() string {
			if r.Name != nil {
				return *r.Name
			}
			return "unknown"
		}()),
	)

	return r, nil
}

func (rs *relayService) GetRelays(
	ctx context.Context,
	filters *RelayFilters,
) ([]domain.Relay, error) {
	if filters == nil {
		filters = &RelayFilters{} // Default empty filters
	}

	opts := repository.ListOption{
		Limit:  filters.Limit,
		Offset: filters.Offset,
		URLs:   filters.URLs,
	}

	// Safe structured logging with nil checks
	logAttrs := []slog.Attr{
		slog.Int("url_count", len(filters.URLs)),
	}

	if filters.Limit != nil {
		logAttrs = append(logAttrs, slog.Int("limit", *filters.Limit))
	}
	if filters.Offset != nil {
		logAttrs = append(logAttrs, slog.Int("offset", *filters.Offset))
	}

	rs.logger.LogAttrs(ctx, slog.LevelInfo, "Fetching relays", logAttrs...)

	relays, err := rs.relayRepo.List(ctx, &opts)
	if err != nil {
		rs.logger.Error("Failed to fetch relays", slog.String("error", err.Error()))
		return nil, fmt.Errorf("could not find relays from the relay repository: %w", err)
	}

	rs.logger.Info("Successfully fetched relays", slog.Int("count", len(relays)))

	return relays, nil
}
