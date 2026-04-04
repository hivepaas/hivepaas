-- +migrate Up
CREATE TABLE IF NOT EXISTS system_statuses
(
    id                      SMALLINT PRIMARY KEY,
    installation_complete   BOOL NOT NULL DEFAULT FALSE,
    update_ver              INT4 NOT NULL DEFAULT 1,

    created_at   TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at   TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP
);

INSERT INTO system_statuses (id) VALUES (1);

-- +migrate Down
DROP TABLE IF EXISTS system_statuses;
