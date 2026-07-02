-- +migrate Up
CREATE TABLE IF NOT EXISTS sys_errors
(
    id           VARCHAR(100) PRIMARY KEY,
    request_id   VARCHAR NULL,
    status       INT NOT NULL,
    code         VARCHAR NULL,
    detail       VARCHAR NULL,
    cause        VARCHAR NULL,
    debug_log    VARCHAR NULL,
    stack_trace  VARCHAR NULL,

    created_at   TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_sys_errors_created_at ON sys_errors(created_at);
CREATE INDEX idx_sys_errors_status ON sys_errors(status);

-- +migrate Down
DROP TABLE IF EXISTS sys_errors;
