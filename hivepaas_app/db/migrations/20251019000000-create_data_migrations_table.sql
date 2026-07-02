-- +migrate Up
CREATE TABLE IF NOT EXISTS data_migrations
(
    id          VARCHAR(100) PRIMARY KEY,
    applied_at  TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_data_migrations_applied_at ON data_migrations(applied_at);

INSERT INTO data_migrations(id) VALUES ('v1-20251019');

-- +migrate Down
DROP TABLE IF EXISTS data_migrations;
