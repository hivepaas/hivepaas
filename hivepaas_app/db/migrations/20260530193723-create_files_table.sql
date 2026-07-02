-- +migrate Up
CREATE TABLE IF NOT EXISTS files
(
    id              VARCHAR(100) PRIMARY KEY,
    scope           VARCHAR(50) NOT NULL,
    object_id       VARCHAR(100) NULL,
    type            VARCHAR(100) NOT NULL,
    kind            VARCHAR(100) NULL,
    key             VARCHAR(100) NULL,
    status          VARCHAR NOT NULL CONSTRAINT chk_status CHECK
                        (status IN ('active','pending','disabled','deleting')) DEFAULT 'active',
    name            VARCHAR(100) NOT NULL,
    path            VARCHAR(100) NOT NULL,
    size            BIGINT NOT NULL,
    mimetype        VARCHAR(100) NULL,
    storage_type    VARCHAR(100) NOT NULL,
    storage_id      VARCHAR(100) NULL,
    bucket          VARCHAR(200) NULL,
    deleted         BOOL NOT NULL DEFAULT FALSE,
    update_ver      INT4 NOT NULL DEFAULT 1,

    created_at   TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at   TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    deleted_at   TIMESTAMPTZ NULL
);

CREATE INDEX idx_files_scope ON files(scope);
CREATE INDEX idx_files_object_id ON files(object_id);
CREATE INDEX idx_files_type ON files(type);
CREATE INDEX idx_files_kind ON files(kind);
CREATE INDEX idx_files_key ON files(key);
CREATE INDEX idx_files_status ON files(status);
CREATE INDEX idx_files_name ON files(name);
CREATE INDEX idx_files_storage_type ON files(storage_type);

CREATE INDEX idx_files_updated_at ON files(updated_at);
CREATE INDEX idx_files_deleted_at ON files(deleted_at);

-- +migrate Down
DROP TABLE IF EXISTS files;
