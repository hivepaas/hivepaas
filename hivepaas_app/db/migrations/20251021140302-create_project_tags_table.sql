-- +migrate Up
CREATE TABLE IF NOT EXISTS project_tags
(
    project_id    VARCHAR(100) NOT NULL,
    tag           VARCHAR(255) NOT NULL,
    display_order INT2 NOT NULL,
    deleted_at    TIMESTAMPTZ NULL,

    PRIMARY KEY (project_id, tag),
    CONSTRAINT fk_project_tags_project_id FOREIGN KEY (project_id) REFERENCES projects (id)
);

-- +migrate Down
DROP TABLE IF EXISTS project_tags;
