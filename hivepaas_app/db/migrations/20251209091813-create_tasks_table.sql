-- +migrate Up
CREATE TABLE IF NOT EXISTS tasks
(
    id               VARCHAR(100) PRIMARY KEY,
    scope            VARCHAR(50) NOT NULL,
    object_id        VARCHAR(100) NULL,
    target_id        VARCHAR(100) NULL,
    type             VARCHAR(100) NOT NULL,
    status           VARCHAR NOT NULL CONSTRAINT chk_status CHECK
                        (status IN ('not-started','in-progress','canceled','failed','done')),
    config           JSONB NOT NULL,
    args             JSONB NULL,
    runs             JSON NULL,
    output           JSON NULL,
    version          INT2 NOT NULL DEFAULT 1,
    update_ver       INT4 NOT NULL DEFAULT 1,

    run_at           TIMESTAMPTZ NULL,
    retry_at         TIMESTAMPTZ NULL,
    started_at       TIMESTAMPTZ NULL,
    ended_at         TIMESTAMPTZ NULL,

    created_at       TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at       TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    deleted_at       TIMESTAMPTZ NULL
);

CREATE UNIQUE INDEX idx_uq_tasks ON tasks(target_id, run_at) WHERE target_id IS NOT NULL AND deleted_at IS NULL;

CREATE INDEX idx_tasks_scope ON tasks(scope);
CREATE INDEX idx_tasks_object_id ON tasks(object_id);
CREATE INDEX idx_tasks_target_id ON tasks(target_id);
CREATE INDEX idx_tasks_type ON tasks(type);
CREATE INDEX idx_tasks_status ON tasks(status);
CREATE INDEX idx_tasks_run_at ON tasks(run_at);
CREATE INDEX idx_tasks_created_at ON tasks(created_at);
CREATE INDEX idx_tasks_updated_at ON tasks(updated_at);
CREATE INDEX idx_tasks_deleted_at ON tasks(deleted_at);

-- +migrate Down
DROP TABLE IF EXISTS tasks;
