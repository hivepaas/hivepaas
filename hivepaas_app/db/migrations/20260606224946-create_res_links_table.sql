-- +migrate Up
CREATE TABLE IF NOT EXISTS res_links
(
    src_type        VARCHAR(100) NOT NULL,
    src_id          VARCHAR NOT NULL,
    dst_type        VARCHAR(100) NOT NULL,
    dst_id          VARCHAR NOT NULL,
    index           INT2 NOT NULL DEFAULT 0,
    data            JSONB NULL,

    created_at   TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at   TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    deleted_at   TIMESTAMPTZ NULL,

    PRIMARY KEY (src_type, src_id, dst_type, dst_id)
);

-- CREATE INDEX idx_res_links_dst_type ON res_links(dst_type);
CREATE INDEX idx_res_links_dst_id ON res_links(dst_id);
CREATE INDEX idx_res_links_updated_at ON res_links(updated_at);
CREATE INDEX idx_res_links_deleted_at ON res_links(deleted_at);

-- +migrate Down
DROP TABLE IF EXISTS res_links;
