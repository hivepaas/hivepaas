-- +migrate Up
-- Where a scan read its image: on the node that runs it, or from its registry,
-- when the node the scans run on is not that one.
ALTER TABLE image_scans ADD COLUMN IF NOT EXISTS source VARCHAR(20) NULL;

-- +migrate Down
ALTER TABLE image_scans DROP COLUMN IF EXISTS source;
