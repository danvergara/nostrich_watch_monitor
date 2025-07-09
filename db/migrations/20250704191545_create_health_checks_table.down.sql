-- Drop indexes first (before dropping the table)
DROP INDEX IF EXISTS idx_health_checks_relay_url;
DROP INDEX IF EXISTS idx_health_checks_relay_created_at;
DROP INDEX IF EXISTS idx_health_checks_relay_time_success;

-- Drop the table
DROP TABLE IF EXISTS health_checks;
