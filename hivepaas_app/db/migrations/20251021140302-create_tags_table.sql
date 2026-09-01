-- +migrate Up
CREATE TABLE IF NOT EXISTS tags
(
    object_id    VARCHAR(100) NOT NULL,
    tag          VARCHAR(255) NOT NULL,
    index        INT2 NOT NULL,
    deleted_at   TIMESTAMPTZ NULL,

    PRIMARY KEY (object_id, tag)
);

CREATE INDEX IF NOT EXISTS idx_tags_tag ON tags(tag);

-- +migrate Down
DROP TABLE IF EXISTS tags;
