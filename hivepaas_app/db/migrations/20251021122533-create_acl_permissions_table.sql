-- +migrate Up
CREATE TABLE IF NOT EXISTS acl_permissions
(
    subj_type   VARCHAR(100) NOT NULL,
    subj_id     VARCHAR(100) NOT NULL,
    res_type    VARCHAR(100) NOT NULL,
    res_id      VARCHAR(100) NOT NULL,
    p_read      BOOL NOT NULL DEFAULT FALSE,
    p_exec      BOOL NOT NULL DEFAULT FALSE,
    p_write     BOOL NOT NULL DEFAULT FALSE,
    p_del       BOOL NOT NULL DEFAULT FALSE,

    created_at     TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at     TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    deleted_at     TIMESTAMPTZ NULL,

    PRIMARY KEY (subj_id, res_id)
);

-- no need separate index on subj_id as it is the leftmost column of the primary key
CREATE INDEX idx_acl_permissions_res_id ON acl_permissions(res_id);
CREATE INDEX idx_acl_permissions_deleted_at ON acl_permissions(deleted_at);

-- +migrate Down
DROP TABLE IF EXISTS acl_permissions;
