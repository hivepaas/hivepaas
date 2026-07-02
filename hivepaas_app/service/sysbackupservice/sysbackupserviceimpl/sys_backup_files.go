package sysbackupserviceimpl

import (
	"archive/tar"
	"context"
	"io"
	"os"
	"path/filepath"

	"github.com/hivepaas/hivepaas/hivepaas_app/apperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/tasklog"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/timeutil"
)

var (
	sysBackupFileModels = []*sysBackupFileModel{}
)

type sysBackupFileModel struct {
	Type          string
	PageDataSize  int
	DirPath       func() string
	TargetDirPath string
}

func (s *service) sysBackupFiles(
	ctx context.Context,
	tarW *tar.Writer,
	data *sysBackupData,
) (err error) {
	start := timeutil.NowUTC()
	_ = data.LogStore.Add(ctx, tasklog.NewWarnFrame("Start backing up static files...", tasklog.TsNow))

	defer func() {
		duration := timeutil.NowUTC().Sub(start)
		if err != nil {
			_ = data.LogStore.Add(ctx, tasklog.NewWarnFrame("Files backup finished in "+duration.String()+
				" with error: "+err.Error(), tasklog.TsNow))
		} else {
			_ = data.LogStore.Add(ctx, tasklog.NewOutFrame("Files backup finished in "+duration.String(),
				tasklog.TsNow))
		}
	}()

	for _, model := range sysBackupFileModels {
		dirPath := model.DirPath()
		targetDirPath := model.TargetDirPath
		entries, err := os.ReadDir(dirPath)
		if err != nil {
			return apperrors.New(err)
		}
		for _, entry := range entries {
			if entry.IsDir() {
				continue
			}
			fileName := entry.Name()
			fileInfo, err := entry.Info()
			if err != nil {
				return apperrors.New(err)
			}

			header, err := tar.FileInfoHeader(fileInfo, "")
			if err != nil {
				return apperrors.New(err)
			}
			header.Name = filepath.ToSlash(filepath.Join(targetDirPath, fileName))

			if err := tarW.WriteHeader(header); err != nil {
				return apperrors.New(err)
			}

			file, err := os.Open(filepath.Join(dirPath, fileName))
			if err != nil {
				return apperrors.New(err)
			}

			if _, err := io.Copy(tarW, file); err != nil {
				file.Close()
				return apperrors.New(err)
			}
			file.Close()
		}
	}

	return nil
}
