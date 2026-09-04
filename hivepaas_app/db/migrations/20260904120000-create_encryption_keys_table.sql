-- +migrate Up
-- The data encryption key that every stored secret is encrypted with, kept here
-- wrapped by the app secret. Changing the app secret re-wraps this row instead of
-- re-encrypting every setting.
CREATE TABLE IF NOT EXISTS encryption_keys
(
    id           VARCHAR(100) PRIMARY KEY,
    wrapped_key  TEXT NOT NULL,
    is_active    BOOL NOT NULL DEFAULT FALSE,

    created_at   TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at   TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP
);

-- Exactly one key may be active. Rotating the key itself, later on, adds a row
-- and moves the flag rather than overwriting anything.
CREATE UNIQUE INDEX idx_encryption_keys_active ON encryption_keys(is_active) WHERE is_active;

-- +migrate Down
DROP TABLE IF EXISTS encryption_keys;
