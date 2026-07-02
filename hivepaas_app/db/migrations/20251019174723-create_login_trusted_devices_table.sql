-- +migrate Up
CREATE TABLE IF NOT EXISTS login_trusted_devices
(
    user_id     VARCHAR(100) NOT NULL,
    device_id   VARCHAR(255) NOT NULL,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at  TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,

    PRIMARY KEY (user_id, device_id),
    CONSTRAINT fk_login_trusted_devices_user_id FOREIGN KEY (user_id) REFERENCES users (id)
);

-- +migrate Down
DROP TABLE IF EXISTS login_trusted_devices;
