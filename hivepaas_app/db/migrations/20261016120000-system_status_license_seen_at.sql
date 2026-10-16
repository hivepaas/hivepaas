-- +migrate Up
-- The latest time license checks have seen. A clock set back more than a day
-- is read as this time, so a key does not hold again by turning the clock back.
ALTER TABLE system_statuses ADD COLUMN IF NOT EXISTS license_seen_at TIMESTAMPTZ NULL;

-- +migrate Down
ALTER TABLE system_statuses DROP COLUMN IF EXISTS license_seen_at;
