-- +migrate Up
CREATE TABLE IF NOT EXISTS app_tags
(
    app_id        VARCHAR(100) NOT NULL,
    tag           VARCHAR(255) NOT NULL,
    display_order INT2 NOT NULL,
    deleted_at    TIMESTAMPTZ NULL,

    PRIMARY KEY (app_id, tag),
    CONSTRAINT fk_app_tags_app_id FOREIGN KEY (app_id) REFERENCES apps (id)
);

-- +migrate Down
DROP TABLE IF EXISTS app_tags;
