package sysbackupserviceimpl

import (
	"context"
	"os/exec"
	"strconv"

	"github.com/hivepaas/hivepaas/hivepaas_app/config"
	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/reflectutil"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/tasklog"
)

// dumpDB dumps HivePaaS's database into path, in pg_dump's custom format, which
// pg_restore reads selectively. The migrations table is left out: a restore runs
// against a database the migrations have made.
func dumpDB(ctx context.Context, path string, logStore *tasklog.Store) error {
	dbConf := config.Current().DB
	pgDumpBin, err := exec.LookPath("pg_dump")
	if err != nil {
		return hperrors.Wrap(err)
	}

	cmd := exec.CommandContext(ctx, pgDumpBin,
		"-h", dbConf.Host,
		"-p", strconv.Itoa(dbConf.Port),
		"-U", dbConf.User,
		"-T", "migrations",
		"-Fc",
		"-f", path,
		dbConf.DBName,
	)
	cmd.Env = []string{"PGPASSWORD=" + dbConf.Password} // NOTE: not use other process's env

	out, err := cmd.CombinedOutput()
	logCmdOutput(ctx, reflectutil.UnsafeBytesToStr(out), err != nil, logStore)
	return hperrors.Wrap(err)
}
