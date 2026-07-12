package sysbackupserviceimpl

import (
	"archive/tar"
	"context"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"

	"github.com/hivepaas/hivepaas/hivepaas_app/apperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/config"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/reflectutil"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/tasklog"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/timeutil"
)

func (s *service) sysBackupDB(
	ctx context.Context,
	tarW *tar.Writer,
	data *sysBackupData,
) (err error) {
	start := timeutil.NowUTC()
	_ = data.LogStore.Add(ctx, tasklog.NewWarnFrame("Start backing up data from DB...", tasklog.TsNow))

	defer func() {
		duration := timeutil.NowUTC().Sub(start)
		if err != nil {
			_ = data.LogStore.Add(ctx, tasklog.NewWarnFrame("DB backup finished in "+duration.String()+
				" with error: "+err.Error(), tasklog.TsNow))
		} else {
			_ = data.LogStore.Add(ctx, tasklog.NewOutFrame("DB backup finished in "+duration.String(),
				tasklog.TsNow))
		}
	}()

	dbConf := config.Current.DB
	dumpFileName := "db.pg_dump"
	dumpFilePath := filepath.Join(data.TempDir, dumpFileName)

	pgDumpBin, err := exec.LookPath("pg_dump")
	if err != nil {
		return apperrors.Wrap(err)
	}

	cmd := exec.CommandContext(ctx, pgDumpBin,
		"-h", dbConf.Host,
		"-p", strconv.Itoa(dbConf.Port),
		"-U", dbConf.User,
		"-T", "migrations",
		"-f", dumpFilePath,
		dbConf.DBName,
	)
	cmd.Env = []string{"PGPASSWORD=" + dbConf.Password} // NOTE: not use other process's env

	out, err := cmd.CombinedOutput()
	logCmdOutput(ctx, reflectutil.UnsafeBytesToStr(out), err != nil, data.LogStore)
	if err != nil {
		return apperrors.Wrap(err)
	}

	fileInfo, err := os.Stat(dumpFilePath)
	if err != nil {
		return apperrors.Wrap(err)
	}

	header, err := tar.FileInfoHeader(fileInfo, "")
	if err != nil {
		return apperrors.Wrap(err)
	}
	header.Name = dumpFileName

	if err := tarW.WriteHeader(header); err != nil {
		return apperrors.Wrap(err)
	}

	file, err := os.Open(dumpFilePath)
	if err != nil {
		return apperrors.Wrap(err)
	}
	defer file.Close()

	if _, err := io.Copy(tarW, file); err != nil {
		return apperrors.Wrap(err)
	}

	return nil
}
