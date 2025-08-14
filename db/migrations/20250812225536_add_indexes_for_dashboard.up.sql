CREATE INDEX idx_relays_name_lower ON relays(LOWER(name));
CREATE INDEX idx_health_checks_latest ON health_checks(relay_url, created_at DESC)
