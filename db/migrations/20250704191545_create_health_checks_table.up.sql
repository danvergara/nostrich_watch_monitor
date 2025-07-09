-- NIP-66 Relay Monitoring - Health Checks Table
CREATE TABLE health_checks (
    id BIGSERIAL PRIMARY KEY,
    
    -- Foreign key to relays table
    relay_url VARCHAR(500) NOT NULL REFERENCES relays(url) ON DELETE CASCADE,
    
    -- Test execution info
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    
    -- WebSocket connection test results (core requirement)
    websocket_success BOOLEAN NOT NULL,
    websocket_error TEXT, -- error message if connection failed
    
    -- NIP-11 document test results (optional metadata gathering)
    nip11_success BOOLEAN, -- nullable since it's optional
    nip11_error TEXT, -- error message if NIP-11 fetch failed
    
    -- RTT measurements (all in milliseconds, nullable if test failed)
    rtt_open INTEGER, -- WebSocket connection establishment time
    rtt_read INTEGER, -- Time to receive data from relay (REQ -> EOSE)
    rtt_write INTEGER, -- Time to publish event and get confirmation
    rtt_nip11 INTEGER -- NIP-11 HTTP fetch time
);

-- Indexes for common queries
CREATE INDEX idx_health_checks_relay_url ON health_checks(relay_url);
CREATE INDEX idx_health_checks_relay_created_at ON health_checks(relay_url, created_at);
CREATE INDEX idx_health_checks_relay_time_success ON health_checks(relay_url, created_at, websocket_success);
