-- +migrate Up
CREATE TABLE IF NOT EXISTS locks
(
    id VARCHAR(200) PRIMARY KEY,
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP
);

INSERT INTO locks (id)
VALUES ('lock:sys:setting-update'),
       ('lock:sys:task-update'),
       ('lock:sys:version-update')
ON CONFLICT DO NOTHING;

-- +migrate Down
DROP TABLE IF EXISTS locks;
