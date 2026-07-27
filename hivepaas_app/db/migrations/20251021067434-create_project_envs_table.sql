-- +migrate Up
CREATE TABLE IF NOT EXISTS project_envs
(
    id           VARCHAR(100) PRIMARY KEY,
    project_id   VARCHAR(100) NOT NULL,
    name         VARCHAR(100) NOT NULL,
    key          VARCHAR(100) NOT NULL,
    status       VARCHAR NOT NULL CONSTRAINT chk_status CHECK
                    (status IN ('active','disabled','deleting')),
    color        VARCHAR(50) NOT NULL,
    index        INT2 NOT NULL DEFAULT 0,
    update_ver   INT4 NOT NULL DEFAULT 1,

    created_at   TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at   TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    deleted_at   TIMESTAMPTZ NULL,

    CONSTRAINT fk_project_envs_project_id FOREIGN KEY (project_id) REFERENCES projects (id)
);

CREATE UNIQUE INDEX idx_uq_project_envs_key ON project_envs(project_id, key) WHERE deleted_at IS NULL;
CREATE INDEX idx_project_envs_status ON project_envs(status);
CREATE INDEX idx_project_envs_updated_at ON project_envs(updated_at);
CREATE INDEX idx_project_envs_deleted_at ON project_envs(deleted_at);

-- +migrate Down
DROP TABLE IF EXISTS project_envs;
