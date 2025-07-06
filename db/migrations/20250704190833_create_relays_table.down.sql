-- Drop indexes first (before dropping the table)
DROP INDEX IF EXISTS idx_relays_created_at;
DROP INDEX IF EXISTS idx_relays_updated_at;
DROP INDEX IF EXISTS idx_relays_supported_nips;
DROP INDEX IF EXISTS idx_relays_tags;
DROP INDEX IF EXISTS idx_relays_countries;

-- Drop the table
DROP TABLE IF EXISTS relays;
