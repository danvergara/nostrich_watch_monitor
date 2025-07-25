package healthcheck

import (
	"context"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/jmoiron/sqlx"
	"github.com/nbd-wtf/go-nostr"
	"github.com/stretchr/testify/require"
)

func TestNewRelayChecker(t *testing.T) {
	db, _, err := sqlmock.New()
	if err != nil {
		t.Fatalf("failed to create mock db: %v", err)
	}
	defer func() {
		_ = db.Close()
	}()

	sqlxDB := sqlx.NewDb(db, "postgres")
	timeout := 30 * time.Second
	privateKey := "test-private-key"
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))

	checker := NewRelayChecker(
		WithDB(sqlxDB),
		WithTimeout(timeout),
		WithPrivateKey(privateKey),
		WithLogger(logger),
	)

	if checker == nil {
		t.Fatal("expected RelayChecker instance, got nil")
	}

	if checker.db != sqlxDB {
		t.Error("database not set correctly")
	}

	if checker.timeout != timeout {
		t.Error("timeout not set correctly")
	}

	if checker.privateKey != privateKey {
		t.Error("private key not set correctly")
	}

	if checker.logger != logger {
		t.Error("logger not set correctly")
	}
}

func TestTestConnectionWithMockWebSocket(t *testing.T) {
	// Create a mock WebSocket server that simulates a Nostr relay
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Simulate WebSocket upgrade headers
		w.Header().Set("Upgrade", "websocket")
		w.Header().Set("Connection", "Upgrade")
		w.Header().Set("Sec-WebSocket-Accept", "test-accept")
		w.WriteHeader(http.StatusSwitchingProtocols)
	}))
	defer server.Close()

	// Convert HTTP URL to WebSocket URL
	wsURL := strings.Replace(server.URL, "http://", "ws://", 1)

	db, _, err := sqlmock.New()
	if err != nil {
		t.Fatalf("failed to create mock db: %v", err)
	}
	defer func() {
		_ = db.Close()
	}()

	sqlxDB := sqlx.NewDb(db, "postgres")
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))

	checker := NewRelayChecker(
		WithDB(sqlxDB),
		WithTimeout(30*time.Second),
		WithPrivateKey("test-key"),
		WithLogger(logger),
	)
	checker.hc = &HealthCheck{
		RelayURL:  wsURL,
		CreatedAt: time.Now(),
	}

	ctx := context.Background()

	// Note: This test will likely fail with the real go-nostr library
	// because our mock server doesn't implement the full WebSocket protocol
	// This demonstrates the test structure for when proper mocking is implemented
	_, err = checker.testConnection(ctx, 5*time.Second)

	// We expect this to fail with our simple mock, but we can verify
	// that the error handling works correctly
	if err == nil {
		t.Log("Connection succeeded (unexpected with simple mock)")
	} else {
		t.Logf("Connection failed as expected with mock server: %v", err)

		// Verify error was recorded
		if checker.hc.WebSocketSuccess {
			t.Error("expected WebSocketSuccess to be false on error")
		}

		if checker.hc.WebSocketError == "" {
			t.Error("expected WebSocketError to be set on error")
		}
	}
}

func TestTestNIP11WithMockServer(t *testing.T) {
	// Create a mock HTTP server that returns NIP-11 JSON
	nip11Response := `{
		"name": "Test Relay",
		"description": "A test relay for unit testing",
		"pubkey": "test-pubkey",
		"contact": "test@example.com",
		"supported_nips": [1, 2, 11],
		"software": "test-software",
		"version": "1.0.0"
	}`

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// NIP-11 requests should have Accept: application/nostr+json header
		if r.Header.Get("Accept") == "application/nostr+json" {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(nip11Response))
		} else {
			w.WriteHeader(http.StatusBadRequest)
		}
	}))
	defer server.Close()

	db, _, err := sqlmock.New()
	if err != nil {
		t.Fatalf("failed to create mock db: %v", err)
	}
	defer func() {
		_ = db.Close()
	}()

	sqlxDB := sqlx.NewDb(db, "postgres")
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))

	checker := NewRelayChecker(
		WithDB(sqlxDB),
		WithTimeout(30*time.Second),
		WithPrivateKey("test-key"),
		WithLogger(logger),
	)
	checker.hc = &HealthCheck{
		RelayURL:  server.URL,
		CreatedAt: time.Now(),
	}

	ctx := context.Background()

	info, err := checker.testNIP11(ctx, 5*time.Second)

	if err != nil {
		t.Fatalf("testNIP11 failed: %v", err)
	}

	// Verify the response was parsed correctly
	if info.Name != "Test Relay" {
		t.Errorf("expected name 'Test Relay', got '%s'", info.Name)
	}

	if info.Description != "A test relay for unit testing" {
		t.Errorf("expected description 'A test relay for unit testing', got '%s'", info.Description)
	}

	// Verify RTT was recorded
	if checker.hc.RTTNIP11 == nil {
		t.Error("expected RTTNIP11 to be set")
	} else if *checker.hc.RTTNIP11 < 0 {
		t.Error("expected positive RTT value")
	}

	// Verify success flag was set
	if !checker.hc.NIP11Success {
		t.Error("expected NIP11Success to be true")
	}
}

func TestTestNIP11WithServerError(t *testing.T) {
	// Create a mock HTTP server that returns an error
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		_, _ = w.Write([]byte("Internal Server Error"))
	}))
	defer server.Close()

	db, _, err := sqlmock.New()
	if err != nil {
		t.Fatalf("failed to create mock db: %v", err)
	}
	defer func() {
		_ = db.Close()
	}()

	sqlxDB := sqlx.NewDb(db, "postgres")
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))

	checker := NewRelayChecker(
		WithDB(sqlxDB),
		WithTimeout(30*time.Second),
		WithPrivateKey("test-key"),
		WithLogger(logger),
	)
	checker.hc = &HealthCheck{
		RelayURL:  server.URL,
		CreatedAt: time.Now(),
	}

	ctx := context.Background()

	_, err = checker.testNIP11(ctx, 5*time.Second)

	if err == nil {
		t.Fatal("expected testNIP11 to fail with server error")
	}

	// Verify error was recorded
	if checker.hc.NIP11Error == "" {
		t.Error("expected NIP11Error to be set on error")
	}
}

func TestAddSupportedNIPs(t *testing.T) {
	type test struct {
		name          string
		supportedNIPs []int
		tags          nostr.Tags
		expectedTags  nostr.Tags
	}

	var tests = []test{
		{
			name:          "success adding nips",
			supportedNIPs: []int{30, 40, 42},
			tags: nostr.Tags{
				{"d", "wss://some.relay/"},
			},
			expectedTags: nostr.Tags{
				{"d", "wss://some.relay/"},
				{"N", "30"},
				{"N", "40"},
				{"N", "42"},
			},
		},
		{
			name:          "empty list of supprted NIPs",
			supportedNIPs: []int{},
			tags: nostr.Tags{
				{"d", "wss://some.relay/"},
			},
			expectedTags: nostr.Tags{
				{"d", "wss://some.relay/"},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			out := addSupportedNIPs(tt.tags, tt.supportedNIPs)
			require.EqualValues(t, tt.expectedTags, out)
		})
	}
}

func TestAddLanguageTags(t *testing.T) {
	type test struct {
		name         string
		languageTags []string
		tags         nostr.Tags
		expectedTags nostr.Tags
	}
	var tests = []test{
		{
			name:         "multiple languages",
			languageTags: []string{"es", "en", "en-419"},
			tags: nostr.Tags{
				{"d", "wss://some.relay/"},
			},
			expectedTags: nostr.Tags{
				{"d", "wss://some.relay/"},
				{"l", "es", "ISO-639-1"},
				{"l", "en", "ISO-639-1"},
				{"l", "en-419", "BCP-47"},
			},
		},
		{
			name:         "global relay",
			languageTags: []string{"*"},
			tags: nostr.Tags{
				{"d", "wss://some.relay/"},
			},
			expectedTags: nostr.Tags{
				{"d", "wss://some.relay/"},
				{"l", "*", "BCP-47"},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			out := addLanguages(tt.tags, tt.languageTags)
			require.EqualValues(t, tt.expectedTags, out)
		})
	}
}
