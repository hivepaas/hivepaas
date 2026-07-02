-- +migrate Up
CREATE TABLE IF NOT EXISTS deployments
(
    id               VARCHAR(100) PRIMARY KEY,
    app_id           VARCHAR(100) NULL,
    status           VARCHAR NOT NULL CONSTRAINT chk_status CHECK
                        (status IN ('not-started','in-progress','canceled','failed','done')) DEFAULT 'not-started',
    settings         JSON NOT NULL,
    trigger          JSONB NOT NULL,
    output           JSON NULL,
    version          INT2 NOT NULL DEFAULT 1,
    update_ver       INT4 NOT NULL DEFAULT 1,

    started_at       TIMESTAMPTZ NULL,
    ended_at         TIMESTAMPTZ NULL,
    created_at       TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at       TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    deleted_at       TIMESTAMPTZ NULL,

    CONSTRAINT fk_deployments_app_id FOREIGN KEY (app_id) REFERENCES apps (id)
);

CREATE INDEX idx_deployments_app_id ON deployments(app_id);
CREATE INDEX idx_deployments_status ON deployments(status);
CREATE INDEX idx_deployments_created_at ON deployments(created_at);
CREATE INDEX idx_deployments_updated_at ON deployments(updated_at);
CREATE INDEX idx_deployments_deleted_at ON deployments(deleted_at);

-- +migrate Down
DROP TABLE IF EXISTS deployments;
