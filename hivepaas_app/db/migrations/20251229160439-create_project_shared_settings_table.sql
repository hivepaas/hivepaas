-- +migrate Up
CREATE TABLE IF NOT EXISTS project_shared_settings
(
    project_id        VARCHAR(100) NOT NULL,
    setting_id        VARCHAR(100) NOT NULL,
    data_view_allowed BOOL NOT NULL DEFAULT FALSE, -- if false, users in project can't see setting data

    created_at   TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    deleted_at   TIMESTAMPTZ NULL,

    PRIMARY KEY(project_id, setting_id),
    CONSTRAINT fk_project_shared_settings_project_id FOREIGN KEY (project_id) REFERENCES projects (id),
    CONSTRAINT fk_project_shared_settings_setting_id FOREIGN KEY (setting_id) REFERENCES settings (id)
);

CREATE INDEX idx_project_shared_settings_setting_id ON project_shared_settings(setting_id);
CREATE INDEX idx_project_shared_settings_created_at ON project_shared_settings(created_at);
CREATE INDEX idx_project_shared_settings_deleted_at ON project_shared_settings(deleted_at);

-- +migrate Down
DROP TABLE IF EXISTS project_shared_settings;
