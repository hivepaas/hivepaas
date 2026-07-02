-- +migrate Up
CREATE TABLE IF NOT EXISTS projects
(
    id           VARCHAR(100) PRIMARY KEY,
    name         VARCHAR(100) NOT NULL,
    key          VARCHAR(100) NOT NULL,
    photo_id     VARCHAR(100) NULL,
    photo        VARCHAR(255) NULL,
    status       VARCHAR NOT NULL CONSTRAINT chk_status CHECK
                    (status IN ('active','disabled','deleting')),
    note         VARCHAR(10000) NULL,
    owner_id     VARCHAR(100) NOT NULL,
    update_ver   INT4 NOT NULL DEFAULT 1,

    created_at   TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at   TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    deleted_at   TIMESTAMPTZ NULL,

    CONSTRAINT fk_project_owner_id FOREIGN KEY (owner_id) REFERENCES users (id)
);

CREATE UNIQUE INDEX idx_uq_projects_name ON projects(LOWER(name)) WHERE deleted_at IS NULL;
CREATE UNIQUE INDEX idx_uq_projects_key ON projects(LOWER(key)) WHERE deleted_at IS NULL;
CREATE INDEX idx_projects_status ON projects(status);
CREATE INDEX idx_projects_updated_at ON projects(updated_at);
CREATE INDEX idx_projects_deleted_at ON projects(deleted_at);

-- +migrate Down
DROP TABLE IF EXISTS projects;
