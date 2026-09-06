-- +migrate Up
CREATE TABLE IF NOT EXISTS audit_logs
(
    id             VARCHAR(100) PRIMARY KEY,
    type           VARCHAR(50)  NOT NULL,
    source         VARCHAR(50)  NULL,
    result         VARCHAR(50)  NOT NULL,

    actor_id       VARCHAR(100) NOT NULL,
    actor_type     VARCHAR(50)  NOT NULL,
    actor_name     VARCHAR(255) NULL,
    via_api_key    BOOLEAN      NOT NULL DEFAULT FALSE,
    session_uid    VARCHAR(100) NULL,

    res_type       VARCHAR(50)  NULL,
    res_id         VARCHAR(100) NULL,
    res_name       VARCHAR(255) NULL,

    client_ip      VARCHAR(50)  NULL,
    remote_addr    VARCHAR(50)  NULL,
    user_agent     VARCHAR      NULL,
    request_id     VARCHAR(100) NULL,

    detail         VARCHAR      NULL,

    created_at     TIMESTAMPTZ  NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_audit_logs_created_at ON audit_logs(created_at);
CREATE INDEX idx_audit_logs_type_created_at ON audit_logs(type, created_at DESC);
CREATE INDEX idx_audit_logs_actor_id_created_at ON audit_logs(actor_id, created_at DESC);
CREATE INDEX idx_audit_logs_res_id_created_at ON audit_logs(res_id, created_at DESC);

-- +migrate Down
DROP TABLE IF EXISTS audit_logs;
