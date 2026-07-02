-- +migrate Up
CREATE TABLE IF NOT EXISTS users
(
    id              VARCHAR(100) PRIMARY KEY,
    username        VARCHAR(100) NOT NULL,
    email           VARCHAR(255) NULL,
    role            VARCHAR NOT NULL CONSTRAINT chk_role CHECK
                        (role IN ('admin','member')),
    status          VARCHAR NOT NULL CONSTRAINT chk_status CHECK
                        (status IN ('active','pending','disabled')),
    full_name       VARCHAR(100) NOT NULL DEFAULT '',
    position        VARCHAR(100) NULL,
    photo           VARCHAR(255) NULL,
    photo_id        VARCHAR(100) NULL,
    notes           VARCHAR(10000) NULL,

    security_option VARCHAR NOT NULL CONSTRAINT chk_security_option CHECK
                        (security_option IN ('enforce-sso','password-2fa','password-only')),
    totp_secret     VARCHAR(100) NULL,
    password        VARCHAR(200) NULL,

    created_at             TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at             TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    access_expire_at       TIMESTAMPTZ NULL,
    last_access            TIMESTAMPTZ NULL,
    deleted_at             TIMESTAMPTZ NULL
);

CREATE UNIQUE INDEX idx_uq_users_username ON users(username) WHERE deleted_at IS NULL;
CREATE UNIQUE INDEX idx_uq_users_email ON users(email) WHERE deleted_at IS NULL;
CREATE INDEX idx_users_full_name ON users(full_name);
CREATE INDEX idx_users_created_at ON users(created_at);
CREATE INDEX idx_users_updated_at ON users(updated_at);
CREATE INDEX idx_users_deleted_at ON users(deleted_at);

-- +migrate Down
DROP TABLE IF EXISTS users;
