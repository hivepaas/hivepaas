-- +migrate Up
CREATE TABLE IF NOT EXISTS apps
(
    id             VARCHAR(100) PRIMARY KEY,
    name           VARCHAR(100) NOT NULL,
    key            VARCHAR(100) NOT NULL,
    global_key     VARCHAR(100) NOT NULL,
    project_id     VARCHAR(100) NOT NULL,
    parent_id      VARCHAR(100) NULL,
    service_id     VARCHAR(100) NULL,
    env            VARCHAR(100) NULL,
    status         VARCHAR NOT NULL CONSTRAINT chk_status CHECK
                        (status IN ('active','disabled','deleting')),
    note           VARCHAR(10000) NULL,
    update_ver     INT4 NOT NULL DEFAULT 1,

    created_at     TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at     TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    deleted_at     TIMESTAMPTZ NULL,

    CONSTRAINT fk_apps_project_id FOREIGN KEY (project_id) REFERENCES projects (id)
);

CREATE UNIQUE INDEX idx_uq_apps_global_key ON apps(global_key) WHERE deleted_at IS NULL;
CREATE INDEX idx_apps_project_id ON apps(project_id);
CREATE INDEX idx_apps_parent_id ON apps(parent_id);
CREATE INDEX idx_apps_env ON apps(env);
CREATE INDEX idx_apps_updated_at ON apps(updated_at);
CREATE INDEX idx_apps_deleted_at ON apps(deleted_at);

-- +migrate Down
DROP TABLE IF EXISTS apps;
