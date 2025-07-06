-- NIP-66 Relay Monitoring - Relays Table (Simplified)
CREATE TABLE relays (
    -- Primary identifier
    url VARCHAR(500) PRIMARY KEY,
    
    -- Core NIP-11 fields
    name VARCHAR(255),
    description TEXT,
    pubkey VARCHAR(64), -- hex pubkey (32 bytes = 64 hex chars)
    contact VARCHAR(255), -- email, nostr address, or other contact
    
    -- Visual and policy URLs
    icon VARCHAR(500), -- URL to relay icon/logo
    banner VARCHAR(500), -- URL to relay banner image
    privacy_policy VARCHAR(500), -- URL to privacy policy
    terms_of_service VARCHAR(500), -- URL to terms of service
    
    -- Relay software info
    software VARCHAR(255), -- relay software name/URL
    version VARCHAR(100), -- software version
    
    -- Capabilities and metadata
    supported_nips INTEGER[], -- array of supported NIP numbers
    relay_countries VARCHAR(10)[], -- array of country codes
    language_tags VARCHAR(10)[], -- array of language codes
    tags VARCHAR(50)[], -- array of general tags
    posting_policy VARCHAR(500), -- URL to posting policy
    
    -- Tracking timestamps
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- Indexes for common queries
CREATE INDEX idx_relays_created_at ON relays(created_at);
CREATE INDEX idx_relays_updated_at ON relays(updated_at);
CREATE INDEX idx_relays_supported_nips ON relays USING GIN(supported_nips);
CREATE INDEX idx_relays_tags ON relays USING GIN(tags);
CREATE INDEX idx_relays_countries ON relays USING GIN(relay_countries);
