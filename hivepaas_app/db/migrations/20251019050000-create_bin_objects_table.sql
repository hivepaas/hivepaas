-- +migrate Up
CREATE TABLE IF NOT EXISTS bin_objects
(
    id              VARCHAR(100) PRIMARY KEY,
    type            VARCHAR(100) NOT NULL,
    status          VARCHAR NOT NULL CONSTRAINT chk_status CHECK
                        (status IN ('active','disabled')),
    name            VARCHAR(200) NULL,
    content_type    VARCHAR(50) NULL,
    data            BYTEA NOT NULL,

    created_at      TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    deleted_at      TIMESTAMPTZ NULL
);

CREATE INDEX idx_bin_objects_type ON bin_objects(type);
CREATE INDEX idx_bin_objects_created_at ON bin_objects(created_at);
CREATE INDEX idx_bin_objects_deleted_at ON bin_objects(deleted_at);

-- +migrate Down
DROP TABLE IF EXISTS bin_objects;
