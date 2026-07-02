-- +migrate Up
CREATE TABLE IF NOT EXISTS task_logs
(
    id              BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    task_id         VARCHAR(100) NOT NULL,
    target_id       VARCHAR(100) NULL,
    type            VARCHAR(255) NULL,
    data            TEXT NOT NULL,
    ts              TIMESTAMPTZ NULL
);

CREATE INDEX idx_task_logs_task_id ON task_logs(task_id);
CREATE INDEX idx_task_logs_target_id ON task_logs(target_id);

-- +migrate Down
DROP TABLE IF EXISTS task_logs;
