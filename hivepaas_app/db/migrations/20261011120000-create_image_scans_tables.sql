-- +migrate Up
-- The last scan of each image, by its ID: the config digest, the same on every
-- node of a platform. Apps on the same image share it.
CREATE TABLE IF NOT EXISTS image_scans
(
    image_id       VARCHAR(100) PRIMARY KEY,
    image_ref      VARCHAR      NOT NULL,
    platform       VARCHAR(50)  NULL,
    os_family      VARCHAR(50)  NULL,
    os_name        VARCHAR(100) NULL,
    os_eosl        BOOLEAN      NOT NULL DEFAULT FALSE,
    status         VARCHAR(20)  NOT NULL,
    error          VARCHAR      NULL,
    trivy_version  VARCHAR(50)  NULL,
    db_updated_at  TIMESTAMPTZ  NULL,
    scanned_at     TIMESTAMPTZ  NOT NULL,
    updated_at     TIMESTAMPTZ  NOT NULL DEFAULT CURRENT_TIMESTAMP
);

-- A scan replaces its image's findings whole.
CREATE TABLE IF NOT EXISTS image_scan_findings
(
    image_id          VARCHAR(100) NOT NULL REFERENCES image_scans (image_id) ON DELETE CASCADE,
    vuln_id           VARCHAR(100) NOT NULL,
    pkg_name          VARCHAR(255) NOT NULL,
    pkg_type          VARCHAR(50)  NOT NULL,
    pkg_path          VARCHAR      NOT NULL DEFAULT '',
    installed_version VARCHAR(255) NOT NULL,
    fixed_version     VARCHAR      NULL,
    severity          VARCHAR(20)  NOT NULL,
    title             VARCHAR      NULL,
    url               VARCHAR      NULL,
    published_at      TIMESTAMPTZ  NULL,
    PRIMARY KEY (image_id, vuln_id, pkg_name, pkg_path, installed_version)
);

-- "Which images carry CVE-X" is a question by the vulnerability.
CREATE INDEX idx_image_scan_findings_vuln_id ON image_scan_findings (vuln_id);

-- What each app's tasks ran when it was last scanned.
CREATE TABLE IF NOT EXISTS app_image_scans
(
    app_id      VARCHAR(100) PRIMARY KEY,
    image_ids   VARCHAR(100)[] NOT NULL DEFAULT '{}',
    error       VARCHAR      NULL,
    scanned_at  TIMESTAMPTZ  NOT NULL,
    reported    VARCHAR[]    NULL
);

CREATE INDEX idx_app_image_scans_image_ids ON app_image_scans USING GIN (image_ids);

-- +migrate Down
DROP TABLE IF EXISTS app_image_scans;
DROP TABLE IF EXISTS image_scan_findings;
DROP TABLE IF EXISTS image_scans;
