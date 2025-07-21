package healthcheck

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
	"github.com/nbd-wtf/go-nostr"
	"github.com/nbd-wtf/go-nostr/nip11"

	"github.com/danvergara/nostrich_watch_monitor/pkg/domain"
	"github.com/danvergara/nostrich_watch_monitor/pkg/repository/postgres"
)

// HealthCheck represents a health check result.
type HealthCheck struct {
	RelayURL         string
	WebSocketSuccess bool
	WebSocketError   string
	NIP11Success     bool
	NIP11Error       string
	RTTOpen          *int // milliseconds
	RTTRead          *int // milliseconds
	RTTWrite         *int // milliseconds
	RTTNIP11         *int // milliseconds
	CreatedAt        time.Time
}

// RelayChecker handles health checking for relays.
type RelayChecker struct {
	db         *sqlx.DB
	timeout    time.Duration
	privateKey string
	hc         *HealthCheck
	logger     *slog.Logger
}

// NewRelayChecker returns a RelayChecker instance given the necessary parameters.
// The privateKey is expected to come from a environment variable.
func NewRelayChecker(
	db *sqlx.DB,
	timeout time.Duration,
	privateKey string,
	logger *slog.Logger,
) *RelayChecker {
	return &RelayChecker{
		db:         db,
		timeout:    timeout,
		privateKey: privateKey,
		logger:     logger,
	}
}

// CheckRelay performs a health check on a single relay.
func (rc *RelayChecker) CheckRelay(ctx context.Context, relayURL string) error {
	rc.hc = &HealthCheck{
		RelayURL:  relayURL,
		CreatedAt: time.Now(),
	}

	// Test WebSocket connection and get relay instance.
	_, err := rc.testConnection(ctx, rc.timeout)
	if err != nil {
		return err
	}

	// Test NIP-11 document (optional).
	info, err := rc.testNIP11(ctx, rc.timeout)
	if err != nil {
		rc.logger.Error(
			fmt.Sprintf("❌ failed to get relay info for %s: %v", relayURL, err),
		)
		return err
	}

	// If NIP-11 was successful, update relay metadata.
	supportedNIPs, err := convertAnyToInt(info.SupportedNIPs)
	if err != nil {
		return err
	}

	relayInfo := domain.Relay{
		URL:            relayURL,
		Name:           info.Name,
		Description:    info.Description,
		PubKey:         info.PubKey,
		Contact:        info.Contact,
		SupportedNIPs:  supportedNIPs,
		Software:       info.Software,
		Version:        info.Version,
		Icon:           info.Icon,
		Banner:         info.Banner,
		PostingPolicy:  info.PostingPolicy,
		Tags:           info.Tags,
		LanguageTags:   info.LanguageTags,
		RelayCountries: info.RelayCountries,
	}

	relayRepo := postgres.NewRelayRepository(rc.db)

	if err := relayRepo.Update(ctx, relayInfo); err != nil {
		rc.logger.Error(
			fmt.Sprintf("❌ failed to update relay info for %s: %v", relayURL, err),
		)
		return err
	}

	hc := domain.HealthCheck{
		RelayURL:         rc.hc.RelayURL,
		CreatedAt:        rc.hc.CreatedAt,
		WebsocketSuccess: rc.hc.WebSocketSuccess,
		WebsocketError:   nullString(rc.hc.WebSocketError),
		Nip11Success:     nullBool(rc.hc.NIP11Success),
		Nip11Error:       nullString(rc.hc.NIP11Error),
		RTTOpen:          rc.hc.RTTOpen,
		RTTRead:          rc.hc.RTTRead,
		RTTWrite:         rc.hc.RTTWrite,
		RTTNIP11:         rc.hc.RTTNIP11,
	}

	if err := relayRepo.SaveHealthCheck(ctx, hc); err != nil {
		rc.logger.Error(
			fmt.Sprintf("❌ failed to update health checks for %s: %v", relayURL, err),
		)
		return err
	}

	return nil
}

// testConnection tests connecting to the relay.
func (rc *RelayChecker) testConnection(
	ctx context.Context,
	timeout time.Duration,
) (*nostr.Relay, error) {
	start := time.Now()

	// Create context with timeout.
	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	// Connect to relay using go-nostr library.
	relay, err := nostr.RelayConnect(ctx, rc.hc.RelayURL)
	if err != nil {
		rc.logger.Error(fmt.Sprintf("❌ failed to connect to %s: %v", rc.hc.RelayURL, err))
		rc.hc.WebSocketSuccess = false
		rc.hc.WebSocketError = err.Error()
		return nil, err
	}

	// Calculate RTT.
	rttMs := int(time.Since(start).Milliseconds())
	rc.hc.RTTOpen = &rttMs
	rc.hc.WebSocketSuccess = true

	rc.logger.Info(fmt.Sprintf("✅ Connected to relay %s (RTT: %dms)", rc.hc.RelayURL, rttMs))

	return relay, nil
}

// testNIP11 tests fetching the NIP-11 information document using go-nostr library.
func (rc *RelayChecker) testNIP11(
	ctx context.Context,
	timeout time.Duration,
) (nip11.RelayInformationDocument, error) {
	start := time.Now()

	// Create context with timeout.
	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	// Use go-nostr's NIP-11 package to fetch relay information.
	rc.logger.Info(fmt.Sprintf("getting NIP 11 response from relay %s", rc.hc.RelayURL))
	info, err := nip11.Fetch(ctx, rc.hc.RelayURL)
	if err != nil {
		rc.logger.Error(
			fmt.Sprintf("❌ failed to get NIP 11 response from relay %s, %s", info.URL, err),
		)
		rc.hc.NIP11Error = err.Error()
		return nip11.RelayInformationDocument{}, err
	}

	// Calculate RTT.
	rttMs := int(time.Since(start).Milliseconds())
	rc.hc.RTTNIP11 = &rttMs
	rc.hc.NIP11Success = true

	rc.logger.Info(fmt.Sprintf("✅ NIP-11 fetch successful from %s (RTT: %dms) - Name: %s",
		rc.hc.RelayURL, rttMs, info.Name))

	return info, nil
}
