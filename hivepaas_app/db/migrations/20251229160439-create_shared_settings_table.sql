-- +migrate Up
CREATE TABLE IF NOT EXISTS shared_settings
(
    scope         VARCHAR(50) NOT NULL,
    object_id     VARCHAR(100) NOT NULL,
    setting_id    VARCHAR(100) NOT NULL,
    can_view_data BOOL NOT NULL DEFAULT TRUE, -- if false, users in project can't see setting data

    created_at   TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    deleted_at   TIMESTAMPTZ NULL,

    PRIMARY KEY(object_id, setting_id),
    CONSTRAINT fk_shared_settings_setting_id FOREIGN KEY (setting_id) REFERENCES settings (id)
);

CREATE INDEX idx_shared_settings_setting_id ON shared_settings(setting_id);
CREATE INDEX idx_shared_settings_created_at ON shared_settings(created_at);
CREATE INDEX idx_shared_settings_deleted_at ON shared_settings(deleted_at);

-- +migrate Down
DROP TABLE IF EXISTS shared_settings;
