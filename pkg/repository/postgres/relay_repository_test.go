/*
Package postgres provides comprehensive integration tests for the RelayRepository using testcontainers.

Test Coverage Overview:

LIST METHOD TESTS:
=================
1. TestList_EmptyDatabase
  - Purpose: Verify empty slice returned when no relays exist
  - Scenario: Query empty database
  - Expected: Empty slice, no error

2. TestList_BasicListing
  - Purpose: Test basic relay retrieval without health checks
  - Scenario: Multiple relays, no health check data
  - Expected: All relays returned with nil HealthCheck fields

3. TestList_WithHealthChecks
  - Purpose: Test relay listing with joined health check data
  - Scenario: Relays with corresponding health checks
  - Expected: Relays with populated HealthCheck fields (RTT, success flags)

4. TestList_LatestHealthCheckOnly
  - Purpose: Ensure only the most recent health check per relay is returned
  - Scenario: Single relay with multiple health checks at different times
  - Expected: Only the latest health check data joined to relay

5. TestList_WithPagination
  - Purpose: Test limit/offset functionality for large datasets
  - Scenario: Multiple relays with limit and offset parameters
  - Expected: Correct subset of relays based on pagination params

6. TestList_WithURLFilter
  - Purpose: Test filtering by specific relay URLs
  - Scenario: Multiple relays, filter by subset of URLs
  - Expected: Only relays matching the URL filter returned

7. TestList_ComplexDataTypes
  - Purpose: Test PostgreSQL arrays and custom data types
  - Scenario: Relay with supported_nips array, tags, etc.
  - Expected: Arrays properly serialized/deserialized from database

FINDBYURL METHOD TESTS:
======================
1. TestFindByURL_ExistingRelay
  - Purpose: Test successful retrieval of existing relay with health check
  - Scenario: Relay exists with health check data
  - Expected: Complete relay object with health check populated

2. TestFindByURL_NonExistentRelay
  - Purpose: Test error handling for missing relays
  - Scenario: Query for non-existent relay URL
  - Expected: Appropriate error returned (sql.ErrNoRows wrapped)

3. TestFindByURL_RelayWithoutHealthCheck
  - Purpose: Test relay retrieval when no health checks exist
  - Scenario: Relay exists but no health check records
  - Expected: Relay returned with nil HealthCheck field
  - Note: This test will catch the table name bug ("relay" vs "relays" in FROM clause)

4. TestFindByURL_LatestHealthCheck
  - Purpose: Ensure latest health check is returned when multiple exist
  - Scenario: Relay with multiple health checks at different timestamps
  - Expected: Relay with most recent health check data

TESTING APPROACH:
================
- Uses testcontainers with PostgreSQL 15 for real database testing
- Runs actual migrations using golang-migrate
- Tests complex SQL joins and subqueries against real database
- Validates PostgreSQL-specific features (arrays, custom types)
- Each test starts with clean database state
- Helper functions provide consistent test data seeding

BENEFITS OVER MOCKING:
=====================
- Catches SQL syntax errors and PostgreSQL-specific issues
- Validates complex join logic and subquery behavior
- Tests real data type serialization/deserialization
- Ensures migration compatibility with repository code
- Provides confidence in actual database interactions
*/
package postgres

import (
	"context"
	"fmt"
	"path/filepath"
	"testing"
	"time"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/jmoiron/sqlx"
	"github.com/lib/pq"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/postgres"
	"github.com/testcontainers/testcontainers-go/wait"

	"github.com/danvergara/nostrich_watch_monitor/pkg/domain"
	"github.com/danvergara/nostrich_watch_monitor/pkg/repository"
)

type RelayRepositoryTestSuite struct {
	suite.Suite
	container *postgres.PostgresContainer
	db        *sqlx.DB
	repo      repository.RelayRepository
	ctx       context.Context
}

func (suite *RelayRepositoryTestSuite) SetupSuite() {
	suite.ctx = context.Background()

	// Start PostgreSQL container
	pgContainer, err := postgres.Run(suite.ctx,
		"postgres:15-alpine",
		postgres.WithDatabase("testdb"),
		postgres.WithUsername("testuser"),
		postgres.WithPassword("testpass"),
		testcontainers.WithWaitStrategy(
			wait.ForLog("database system is ready to accept connections").
				WithOccurrence(2).WithStartupTimeout(30*time.Second)),
	)
	require.NoError(suite.T(), err)
	suite.container = pgContainer

	// Get connection details
	host, err := pgContainer.Host(suite.ctx)
	require.NoError(suite.T(), err)
	port, err := pgContainer.MappedPort(suite.ctx, "5432")
	require.NoError(suite.T(), err)

	// Connect to database
	dsn := fmt.Sprintf("postgres://testuser:testpass@%s:%s/testdb?sslmode=disable",
		host, port.Port())

	suite.db, err = sqlx.Connect("postgres", dsn)
	require.NoError(suite.T(), err)

	// Run migrations
	err = suite.runMigrations(dsn)
	require.NoError(suite.T(), err)

	// Create repository
	suite.repo = NewRelayRepository(suite.db)
}

func (suite *RelayRepositoryTestSuite) TearDownSuite() {
	if suite.db != nil {
		suite.db.Close()
	}
	if suite.container != nil {
		suite.container.Terminate(suite.ctx)
	}
}

func (suite *RelayRepositoryTestSuite) SetupTest() {
	// Clean tables before each test
	suite.cleanTables()
}

func (suite *RelayRepositoryTestSuite) runMigrations(dsn string) error {
	// Get absolute path to migrations
	migrationsPath, err := filepath.Abs("../../../db/migrations")
	if err != nil {
		return err
	}

	m, err := migrate.New(
		fmt.Sprintf("file://%s", migrationsPath),
		dsn,
	)
	if err != nil {
		return err
	}
	defer m.Close()

	return m.Up()
}

func (suite *RelayRepositoryTestSuite) cleanTables() {
	// Clean in reverse order due to foreign keys
	suite.db.MustExec("DELETE FROM health_checks")
	suite.db.MustExec("DELETE FROM relays")
}

// Test data helpers
func (suite *RelayRepositoryTestSuite) seedRelay(url string, name string) domain.Relay {
	namePtr := name
	descPtr := fmt.Sprintf("Test relay %s", name)
	pubkeyPtr := "test-pubkey-" + name
	contactPtr := fmt.Sprintf("test-%s@example.com", name)
	softwarePtr := "test-software"
	versionPtr := "1.0.0"

	relay := domain.Relay{
		URL:           url,
		Name:          &namePtr,
		Description:   &descPtr,
		PubKey:        &pubkeyPtr,
		Contact:       &contactPtr,
		SupportedNIPs: pq.Int64Array{1, 2, 11},
		Software:      &softwarePtr,
		Version:       &versionPtr,
	}

	query := `
        INSERT INTO relays (url, name, description, pubkey, contact, supported_nips, software, version)
        VALUES (:url, :name, :description, :pubkey, :contact, :supported_nips, :software, :version)
    `
	_, err := suite.db.NamedExec(query, relay)
	require.NoError(suite.T(), err)

	return relay
}

func (suite *RelayRepositoryTestSuite) seedHealthCheck(
	relayURL string,
	createdAt time.Time,
	websocketSuccess bool,
) domain.HealthCheck {
	hc := domain.HealthCheck{
		RelayURL:         relayURL,
		CreatedAt:        &createdAt,
		WebsocketSuccess: &websocketSuccess,
		Nip11Success:     &[]bool{true}[0],
		RTTOpen:          &[]int{100}[0],
		RTTRead:          &[]int{200}[0],
		RTTWrite:         &[]int{150}[0],
		RTTNIP11:         &[]int{50}[0],
	}

	query := `
        INSERT INTO health_checks (relay_url, created_at, websocket_success, nip11_success, rtt_open, rtt_read, rtt_write, rtt_nip11)
        VALUES (:relay_url, :created_at, :websocket_success, :nip11_success, :rtt_open, :rtt_read, :rtt_write, :rtt_nip11)
    `
	_, err := suite.db.NamedExec(query, hc)
	require.NoError(suite.T(), err)

	return hc
}

// List method tests
func (suite *RelayRepositoryTestSuite) TestList_EmptyDatabase() {
	relays, err := suite.repo.List(suite.ctx, nil)

	require.NoError(suite.T(), err)
	assert.Empty(suite.T(), relays)
}

func (suite *RelayRepositoryTestSuite) TestList_BasicListing() {
	// Seed test data
	suite.seedRelay("wss://relay1.example.com", "Relay 1")
	suite.seedRelay("wss://relay2.example.com", "Relay 2")

	relays, err := suite.repo.List(suite.ctx, nil)

	require.NoError(suite.T(), err)
	assert.Len(suite.T(), relays, 2)
	assert.Equal(suite.T(), "wss://relay1.example.com", relays[0].URL)
	assert.Equal(suite.T(), "wss://relay2.example.com", relays[1].URL)
	assert.Nil(suite.T(), relays[0].HealthCheck) // No health checks yet
}

func (suite *RelayRepositoryTestSuite) TestList_WithHealthChecks() {
	// Seed relays
	suite.seedRelay("wss://relay1.example.com", "Relay 1")
	suite.seedRelay("wss://relay2.example.com", "Relay 2")

	// Seed health checks
	now := time.Now()
	suite.seedHealthCheck("wss://relay1.example.com", now, true)
	suite.seedHealthCheck("wss://relay2.example.com", now.Add(-1*time.Hour), false)

	relays, err := suite.repo.List(suite.ctx, nil)

	require.NoError(suite.T(), err)
	assert.Len(suite.T(), relays, 2)

	// Check first relay has health check
	assert.NotNil(suite.T(), relays[0].HealthCheck)
	assert.True(suite.T(), *relays[0].WebsocketSuccess)
	assert.Equal(suite.T(), 100, *relays[0].RTTOpen)

	// Check second relay has health check
	assert.NotNil(suite.T(), relays[1].HealthCheck)
	assert.False(suite.T(), *relays[1].WebsocketSuccess)
}

func (suite *RelayRepositoryTestSuite) TestList_LatestHealthCheckOnly() {
	// Seed relay
	suite.seedRelay("wss://relay1.example.com", "Relay 1")

	// Seed multiple health checks (older first)
	now := time.Now()
	suite.seedHealthCheck("wss://relay1.example.com", now.Add(-2*time.Hour), false) // Older
	suite.seedHealthCheck("wss://relay1.example.com", now.Add(-1*time.Hour), true)  // Latest

	relays, err := suite.repo.List(suite.ctx, nil)

	require.NoError(suite.T(), err)
	assert.Len(suite.T(), relays, 1)

	// Should have the latest health check (success = true)
	assert.NotNil(suite.T(), relays[0].HealthCheck)
	assert.True(suite.T(), *relays[0].WebsocketSuccess)
}

func (suite *RelayRepositoryTestSuite) TestList_WithPagination() {
	// Seed multiple relays
	for i := 1; i <= 5; i++ {
		suite.seedRelay(fmt.Sprintf("wss://relay%d.example.com", i), fmt.Sprintf("Relay %d", i))
	}

	// Test with limit
	limit := 3
	relays, err := suite.repo.List(suite.ctx, &repository.ListOption{Limit: &limit})

	require.NoError(suite.T(), err)
	assert.Len(suite.T(), relays, 3)

	// Test with offset
	offset := 2
	relays, err = suite.repo.List(suite.ctx, &repository.ListOption{Offset: &offset})

	require.NoError(suite.T(), err)
	assert.Len(suite.T(), relays, 3) // Should get remaining 3 relays
	assert.Equal(suite.T(), "wss://relay3.example.com", relays[0].URL)
}

func (suite *RelayRepositoryTestSuite) TestList_WithURLFilter() {
	// Seed multiple relays
	suite.seedRelay("wss://relay1.example.com", "Relay 1")
	suite.seedRelay("wss://relay2.example.com", "Relay 2")
	suite.seedRelay("wss://relay3.example.com", "Relay 3")

	// Filter by specific URLs
	urls := []string{"wss://relay1.example.com", "wss://relay3.example.com"}
	relays, err := suite.repo.List(suite.ctx, &repository.ListOption{URLs: urls})

	require.NoError(suite.T(), err)
	assert.Len(suite.T(), relays, 2)
	assert.Equal(suite.T(), "wss://relay1.example.com", relays[0].URL)
	assert.Equal(suite.T(), "wss://relay3.example.com", relays[1].URL)
}

func (suite *RelayRepositoryTestSuite) TestList_ComplexDataTypes() {
	// Create relay with complex data types
	namePtr := "Complex Relay"
	relay := domain.Relay{
		URL:           "wss://complex.example.com",
		Name:          &namePtr,
		SupportedNIPs: pq.Int64Array{1, 2, 11, 42, 50},
	}

	query := `
        INSERT INTO relays (url, name, supported_nips)
        VALUES (:url, :name, :supported_nips)
    `
	_, err := suite.db.NamedExec(query, relay)
	require.NoError(suite.T(), err)

	relays, err := suite.repo.List(suite.ctx, nil)

	require.NoError(suite.T(), err)
	assert.Len(suite.T(), relays, 1)
	assert.Equal(suite.T(), pq.Int64Array{1, 2, 11, 42, 50}, relays[0].SupportedNIPs)
}

// FindByURL method tests
func (suite *RelayRepositoryTestSuite) TestFindByURL_ExistingRelay() {
	// Seed relay and health check
	suite.seedRelay("wss://test.example.com", "Test Relay")
	suite.seedHealthCheck("wss://test.example.com", time.Now(), true)

	relay, err := suite.repo.FindByURL(suite.ctx, "wss://test.example.com")

	require.NoError(suite.T(), err)
	assert.Equal(suite.T(), "wss://test.example.com", relay.URL)
	assert.Equal(suite.T(), "Test Relay", *relay.Name)
	assert.NotNil(suite.T(), relay.HealthCheck)
	assert.True(suite.T(), *relay.WebsocketSuccess)
}

func (suite *RelayRepositoryTestSuite) TestFindByURL_NonExistentRelay() {
	_, err := suite.repo.FindByURL(suite.ctx, "wss://nonexistent.example.com")

	assert.Error(suite.T(), err)
	assert.Contains(suite.T(), err.Error(), "not found")
}

func (suite *RelayRepositoryTestSuite) TestFindByURL_RelayWithoutHealthCheck() {
	// Test relay retrieval when no health checks exist
	suite.seedRelay("wss://test.example.com", "Test Relay")

	relay, err := suite.repo.FindByURL(suite.ctx, "wss://test.example.com")

	require.NoError(suite.T(), err)
	assert.Equal(suite.T(), "wss://test.example.com", relay.URL)
	assert.Equal(suite.T(), "Test Relay", *relay.Name)
	// HealthCheck should be nil when no health check data exists
	assert.Nil(suite.T(), relay.HealthCheck)
}

func (suite *RelayRepositoryTestSuite) TestFindByURL_LatestHealthCheck() {
	// Seed relay
	suite.seedRelay("wss://test.example.com", "Test Relay")

	// Seed multiple health checks
	now := time.Now()
	suite.seedHealthCheck("wss://test.example.com", now.Add(-2*time.Hour), false) // Older
	suite.seedHealthCheck("wss://test.example.com", now.Add(-1*time.Hour), true)  // Latest

	relay, err := suite.repo.FindByURL(suite.ctx, "wss://test.example.com")

	require.NoError(suite.T(), err)
	assert.NotNil(suite.T(), relay.HealthCheck)
	assert.True(suite.T(), *relay.WebsocketSuccess) // Should get latest
}

// Run the test suite
func TestRelayRepositoryTestSuite(t *testing.T) {
	suite.Run(t, new(RelayRepositoryTestSuite))
}
