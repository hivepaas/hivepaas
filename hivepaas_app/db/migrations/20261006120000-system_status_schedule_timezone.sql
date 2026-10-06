-- +migrate Up
-- The timezone the daily system jobs were last given their times of day in: a
-- start under another moves those still at theirs. '' is before there was one,
-- when they were given them in UTC.
ALTER TABLE system_statuses ADD COLUMN IF NOT EXISTS schedule_timezone VARCHAR NOT NULL DEFAULT '';

-- +migrate Down
ALTER TABLE system_statuses DROP COLUMN IF EXISTS schedule_timezone;
