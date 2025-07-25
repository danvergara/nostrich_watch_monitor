package healthcheck

import (
	"context"
	"fmt"
	"log/slog"
	"strconv"
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
	db           *sqlx.DB
	timeout      time.Duration
	privateKey   string
	hc           *HealthCheck
	logger       *slog.Logger
	monitorRelay string
}

// Option is a functional option type that allows us to configure the Client.
type Option func(*RelayChecker)

// NewRelayChecker returns a RelayChecker instance given the necessary parameters.
// The privateKey is expected to come from a environment variable.
func NewRelayChecker(options ...Option) *RelayChecker {
	rc := &RelayChecker{}

	// Apply all the functional options to configure the Relay Checker.
	for _, opt := range options {
		opt(rc)
	}

	return rc
}

// WithTimeout is a functional option to set the HTTP Relay Checker timeout.
func WithTimeout(timeout time.Duration) Option {
	return func(rc *RelayChecker) {
		rc.timeout = timeout
	}
}

// WithDB is a functional option to set database pool of connection.
func WithDB(db *sqlx.DB) Option {
	return func(rc *RelayChecker) {
		rc.db = db
	}
}

// WithPrivatekey is a functional option to set monitor's private key.
func WithPrivateKey(privateKey string) Option {
	return func(rc *RelayChecker) {
		rc.privateKey = privateKey
	}
}

// WithPrivatekey is a functional option to set monitor's private key.
func WithLogger(logger *slog.Logger) Option {
	return func(rc *RelayChecker) {
		rc.logger = logger
	}
}

// WithMonitorRelay is a functional option to set monitor's private relay URL.
func WithMonitorRelay(monitorRelay string) Option {
	return func(rc *RelayChecker) {
		rc.monitorRelay = monitorRelay
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

	pub, err := nostr.GetPublicKey(rc.privateKey)
	if err != nil {
		rc.logger.Error(
			fmt.Sprintf("❌ failed to derive the monitor's public key: %v", err),
		)
		return err
	}

	ev := nostr.Event{
		PubKey:    pub,
		CreatedAt: nostr.Now(),
		Kind:      30166,
		Tags: nostr.Tags{
			{"d", relayURL},
			{"n", "clearnet"},
			{"rtt-open", strconv.Itoa(nullInt(rc.hc.RTTOpen))},
		},
		Content: "",
	}

	// Add Supported NIPs to the Tags field.
	ev.Tags = addSupportedNIPs(ev.Tags, supportedNIPs)

	// Add payment and auth requirements, if any.
	ev.Tags = addLimitations(ev.Tags, info.Limitation)

	// Add "Topics" From NIP-11 "Informational Document" nip11.tags[].
	ev.Tags = addTopics(ev.Tags, info.Tags)

	// Add Supported languages by the relay of interest.
	ev.Tags = addLanguages(ev.Tags, info.LanguageTags)

	ev.Sign(rc.privateKey)

	relay, err := nostr.RelayConnect(ctx, rc.monitorRelay)
	if err != nil {
		rc.logger.Error(
			fmt.Sprintf("❌ failed to connect to the monitor's relay: %v", err),
		)
		return err
	}

	if err := relay.Publish(ctx, ev); err != nil {
		rc.logger.Error(
			fmt.Sprintf(
				"❌ failed to publish 30166 event about %s to the monitor's relay: %v",
				relayURL,
				err,
			),
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

func (rc *RelayChecker) Publish10166Event(ctx context.Context, frequency, timeout string) error {
	pub, err := nostr.GetPublicKey(rc.privateKey)
	if err != nil {
		rc.logger.Error(
			fmt.Sprintf("❌ failed to get derive the monitor's public key: %v", err),
		)
		return err
	}

	ev := nostr.Event{
		Kind:      10166,
		PubKey:    pub,
		CreatedAt: nostr.Timestamp(time.Now().Unix()),
		Content:   "",
		Tags: nostr.Tags{
			// Frequency of monitoring (example: every 3600 seconds/1 hour).
			{"frequency", frequency},

			// Checks performed.
			{"c", "ws"},
			{"c", "nip11"},

			// Timeout configurations.
			{"timeout", timeout, "open"},
			{"timeout", timeout, "nip11"},
		},
	}

	// Since it's a replaceable event, it will automatically
	// replace any previous 10166 from this pubkey
	if err = ev.Sign(rc.privateKey); err != nil {
		rc.logger.Error(
			fmt.Sprintf("❌ failed to sign the event using the monitor's private key: %v", err),
		)
		return err
	}

	relay, err := nostr.RelayConnect(ctx, rc.monitorRelay)
	if err != nil {
		rc.logger.Error(
			fmt.Sprintf("❌ failed to connect to the monitor's relay: %v", err),
		)
		return err
	}

	if err := relay.Publish(ctx, ev); err != nil {
		rc.logger.Error(
			fmt.Sprintf(
				"❌ failed to publish 10166 event monitor annnoucement to the monitor's relay: %v",
				err,
			),
		)
		return err
	}

	return nil
}
