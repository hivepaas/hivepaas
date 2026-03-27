-- +migrate Up
CREATE TABLE IF NOT EXISTS settings
(
    id                VARCHAR(100) PRIMARY KEY,
    scope             VARCHAR(50) NOT NULL,
    object_id         VARCHAR(100) NULL,
    type              VARCHAR(100) NOT NULL,
    kind              VARCHAR(100) NULL,
    name              VARCHAR(100) NULL,
    version           INT2 NOT NULL DEFAULT 1,
    status            VARCHAR NOT NULL CONSTRAINT chk_status CHECK
                        (status IN ('active','pending','disabled','expired')) DEFAULT 'active',
    data              JSONB NULL,
    avail_in_projects BOOL NOT NULL DEFAULT FALSE,
    is_default        BOOL NOT NULL DEFAULT FALSE,
    update_ver        INT4 NOT NULL DEFAULT 1,

    created_at   TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at   TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    expire_at    TIMESTAMPTZ NULL,
    deleted_at   TIMESTAMPTZ NULL
);

CREATE INDEX idx_settings_scope ON settings(scope);
CREATE INDEX idx_settings_type ON settings(type);
CREATE INDEX idx_settings_kind ON settings(kind);
CREATE INDEX idx_settings_name ON settings(name);
CREATE INDEX idx_settings_status ON settings(status);
CREATE INDEX idx_settings_object_id ON settings(object_id);
CREATE INDEX idx_settings_updated_at ON settings(updated_at);
CREATE INDEX idx_settings_expire_at ON settings(expire_at);
CREATE INDEX idx_settings_deleted_at ON settings(deleted_at);

-- +migrate Down
DROP TABLE IF EXISTS settings;
