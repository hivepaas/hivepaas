package sysupdateserviceimpl

import (
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"

	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/config"
	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/reflectutil"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/tasklog"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/unit"
)

// dbBackupFileName is a fixed name in a fixed directory, on purpose.
//
// There is exactly one of these and the next update overwrites it, so the disk
// cost is bounded at one dump no matter how many updates an installation goes
// through. It is not deleted on success either: an update that looked fine and
// turns out not to be an hour later is a real thing, and deleting the dump the
// moment the update returns throws away the only copy at precisely the moment it
// becomes interesting.
const dbBackupFileName = "db-backup.dump"

// dbBackupCompression is what pg_dump is asked for, in order of preference.
//
// Measured on a real install's database: 480KB uncompressed, 173KB at zstd:1,
// 186KB at gzip:1, 177KB at gzip:6. zstd:1 beats gzip at its slowest setting on
// both size and speed, so there is no reason to ask for a harder level - this
// runs inside the window where the system is stopped.
//
// gzip is the fallback because the method has to be compiled into pg_dump, and
// finding out it is not - during an update, with the system already down - is
// not a discovery worth making. gzip has been there since long before zstd.
var dbBackupCompression = []string{"zstd:1", "gzip:1"}

// backupDB dumps the whole database to a known path before anything is changed.
//
// Deliberately not sysbackupservice. That one serves the user's own backup
// schedule and excludes the `migrations` table, because it restores data into a
// database whose schema has already been migrated. This one has the opposite
// job: put the database back exactly as it was, sql-migrate's bookkeeping
// included, so a failed migration can be undone.
//
// The custom format is what makes the restore possible at all. A plain SQL dump
// replayed into a database that still has its tables fails on every CREATE and
// appends rather than replaces; pg_restore reads the custom format and can be
// told to drop first, in one transaction.
func (s *service) backupDB(
	ctx context.Context,
	data *sysUpdateData,
) (err error) {
	defer s.logStep(ctx, data, "database backup")(&err)

	path, err := dbBackupPath()
	if err != nil {
		return hperrors.Wrap(err)
	}

	pgDumpBin, err := exec.LookPath("pg_dump")
	if err != nil {
		return hperrors.Wrap(err)
	}

	dbConf := config.Current().DB
	var lastErr error
	for _, compression := range dbBackupCompression {
		cmd := exec.CommandContext(ctx, pgDumpBin, //nolint:gosec // arguments are config, not input
			"-h", dbConf.Host,
			"-p", strconv.Itoa(dbConf.Port),
			"-U", dbConf.User,
			"--format=custom",
			"--compress="+compression,
			"--file="+path,
			dbConf.DBName,
		)
		// Built rather than inherited: the process environment holds the app
		// secret, and pg_dump has no business being handed it.
		cmd.Env = []string{"PGPASSWORD=" + dbConf.Password}

		out, runErr := cmd.CombinedOutput()
		if runErr == nil {
			s.logCmdOutput(ctx, data, out)
			_ = data.LogStore.Add(ctx, tasklog.NewOutFrame(
				"Database backed up to "+path+" ("+compression+", "+dbBackupSize(path)+")", tasklog.TsNow))
			return nil
		}
		lastErr = runErr
		_ = data.LogStore.Add(ctx, tasklog.NewWarnFrame(
			"pg_dump could not use "+compression+", trying the next one", tasklog.TsNow))
		s.logCmdOutput(ctx, data, out)
	}

	return hperrors.Wrap(lastErr)
}

// restoreDB puts the database back from the dump taken before the update.
//
// Called for one failure only: a migration that did not complete. Every other
// step that can fail leaves the database migrated and working, and undoing a
// migration that succeeded to answer a problem with traefik would trade one
// inconsistency for a different one while throwing away work.
//
// --single-transaction is the point of the whole design. A restore that fails
// halfway leaves a database in a state nothing anticipated, which is worse than
// the failed migration it was trying to undo; this way it either lands or it
// does not, and what it does not do is invent a third state.
func (s *service) restoreDB(
	ctx context.Context,
	data *sysUpdateData,
) (err error) {
	defer s.logStep(ctx, data, "database restore")(&err)

	path, err := dbBackupPath()
	if err != nil {
		return hperrors.Wrap(err)
	}
	if _, statErr := os.Stat(path); statErr != nil {
		// Nothing to restore from - the update was told to skip the backup, or
		// the backup itself never ran. Say so rather than failing: the migration
		// error the caller is already returning is the one worth reading.
		_ = data.LogStore.Add(ctx, tasklog.NewWarnFrame(
			"No database backup at "+path+", nothing to restore", tasklog.TsNow))
		return nil
	}

	pgRestoreBin, err := exec.LookPath("pg_restore")
	if err != nil {
		return hperrors.Wrap(err)
	}

	dbConf := config.Current().DB
	cmd := exec.CommandContext(ctx, pgRestoreBin, //nolint:gosec // arguments are config, not input
		"-h", dbConf.Host,
		"-p", strconv.Itoa(dbConf.Port),
		"-U", dbConf.User,
		"--dbname="+dbConf.DBName,
		// Drops what it is about to recreate. Without this the restore lands on
		// top of the objects the failed migration left behind.
		"--clean",
		"--if-exists",
		"--single-transaction",
		path,
	)
	cmd.Env = []string{"PGPASSWORD=" + dbConf.Password}

	out, err := cmd.CombinedOutput()
	s.logCmdOutput(ctx, data, out)
	if err != nil {
		return hperrors.Wrap(err)
	}

	_ = data.LogStore.Add(ctx, tasklog.NewWarnFrame(
		"Database restored from "+path+" - the update is being abandoned", tasklog.TsNow))
	return nil
}

// dbBackupPath is fixed, so an operator can find it without reading a task log
// and the next update knows where to overwrite.
func dbBackupPath() (string, error) {
	dir := config.Current().DataPathSystemUpdate().AbsPath()
	if err := os.MkdirAll(dir, base.DirModeDefault); err != nil {
		return "", hperrors.Wrap(err)
	}
	return filepath.Join(dir, dbBackupFileName), nil
}

func dbBackupSize(path string) string {
	info, err := os.Stat(path)
	if err != nil {
		return "unknown size"
	}
	return unit.DataSize(info.Size()).String()
}

func (s *service) logCmdOutput(ctx context.Context, data *sysUpdateData, out []byte) {
	text := reflectutil.UnsafeBytesToStr(out)
	if text == "" {
		return
	}
	_ = data.LogStore.Add(ctx, tasklog.NewOutFrame(text, tasklog.TsNow))
}
