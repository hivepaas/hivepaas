-- +migrate Up
CREATE TABLE IF NOT EXISTS locks
(
    id VARCHAR(200) PRIMARY KEY,
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP
);

INSERT INTO locks (id)
VALUES ('lock:sys:init-default-settings'),
       ('lock:sys:version-update')
ON CONFLICT DO NOTHING;

-- +migrate Down
DROP TABLE IF EXISTS locks;
