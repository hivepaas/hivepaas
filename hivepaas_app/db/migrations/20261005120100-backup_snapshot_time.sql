-- +migrate Up notransaction
-- A backup snapshot's created_at is when it was taken, as its data's time says:
-- the snapshots are listed newest first, by repository, and an index serves
-- that, where one on the data's time cannot - a text's cast to a timestamp is
-- not immutable. Those recorded so far were given when they were recorded,
-- which for one read from a repository could be long after.
UPDATE settings SET created_at = (data->>'time')::timestamptz
WHERE type = 'backup-snapshot' AND data->>'time' > '1970' AND created_at <> (data->>'time')::timestamptz;

-- The snapshots of a repository, newest first (backupsnapshotuc, query.go):
-- the list asks by these conditions, and must.
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_settings_backup_snapshot_time ON settings (ref_id, created_at DESC)
    WHERE type = 'backup-snapshot' AND deleted_at IS NULL;

-- +migrate Down notransaction
DROP INDEX CONCURRENTLY IF EXISTS idx_settings_backup_snapshot_time;
