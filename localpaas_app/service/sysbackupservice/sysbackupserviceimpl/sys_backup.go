package sysbackupserviceimpl

import (
	"archive/tar"
	"compress/gzip"
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"time"

	"filippo.io/age"
	"github.com/klauspost/compress/zstd"

	"github.com/localpaas/localpaas/localpaas_app/apperrors"
	"github.com/localpaas/localpaas/localpaas_app/base"
	"github.com/localpaas/localpaas/localpaas_app/config"
	"github.com/localpaas/localpaas/localpaas_app/entity"
	"github.com/localpaas/localpaas/localpaas_app/infra/database"
	"github.com/localpaas/localpaas/localpaas_app/pkg/fileutil"
	"github.com/localpaas/localpaas/localpaas_app/pkg/funcutil"
	"github.com/localpaas/localpaas/localpaas_app/pkg/timeutil"
	"github.com/localpaas/localpaas/localpaas_app/service/sysbackupservice"
)

const (
	// 0755 grants read/write/execute for owner, read/execute for group/others
	// 0644 grants read/write for owner, read-only for group/others
	sysBackupDirFileMode = 0o755
)

type sysBackupData struct {
	*sysbackupservice.SysBackupReq
	TaskOutput *entity.TaskSystemBackupOutput
	TimeNow    time.Time

	BackupSaveDir string
	TempDir       string

	OutFileName   string
	OutFilePath   string
	LocalOutFile  *entity.Setting
	RemoteOutFile *entity.Setting
}

func (s *service) Backup(
	ctx context.Context,
	db database.Tx,
	req *sysbackupservice.SysBackupReq,
) (resp *sysbackupservice.SysBackupResp, err error) {
	defer funcutil.EnsureNoPanic(&err)

	resp = &sysbackupservice.SysBackupResp{}
	data := &sysBackupData{
		SysBackupReq: req,
		TaskOutput: &entity.TaskSystemBackupOutput{
			DBBackup: &entity.DBBackupOutput{},
		},
		TimeNow: timeutil.NowUTC(),
	}

	// Backup DB
	err = s.sysBackup(ctx, db, data)
	if err != nil {
		return nil, apperrors.Wrap(err)
	}

	// Assign back the result output
	data.Task.MustSetOutput(data.TaskOutput)

	return resp, nil
}

func (s *service) sysBackup(
	ctx context.Context,
	db database.IDB,
	data *sysBackupData,
) (err error) {
	defer func() {
		if err != nil {
			data.TaskOutput.DBBackup.Error = err.Error()
		}
	}()

	data.TempDir, err = fileutil.CreateTempDirInAppPath("", "sys-backup-*", 0)
	if err != nil {
		return apperrors.Wrap(err)
	}
	defer os.RemoveAll(data.TempDir)

	bakTmpFile, tarW, closer, err := s.sysBackupCreateWriter(data)
	if err != nil {
		return apperrors.Wrap(err)
	}
	defer func() {
		if closer != nil {
			_ = closer()
		}
	}()

	// Start the data backup
	err = s.sysBackupDB(ctx, tarW, data)
	if err != nil {
		return apperrors.Wrap(err)
	}

	err = s.sysBackupFiles(ctx, tarW, data)
	if err != nil {
		return apperrors.Wrap(err)
	}

	_ = closer() // Flush data in writers
	closer = nil

	// Save the result in a local file
	err = s.sysBackupSaveResultInLocal(ctx, db, bakTmpFile, data)
	if err != nil {
		return apperrors.Wrap(err)
	}

	// Upload backup file to cloud storage if configured
	err = s.sysBackupSaveResultInStorage(ctx, db, data)
	if err != nil {
		return apperrors.Wrap(err)
	}

	return nil
}

//nolint:gocognit
func (s *service) sysBackupCreateWriter(
	data *sysBackupData,
) (tmpFileName string, tarW *tar.Writer, closer func() error, err error) {
	// Make sure the backup directory exist
	data.BackupSaveDir = config.Current.DataPathSystemBackupFiles()

	err = os.MkdirAll(data.BackupSaveDir, sysBackupDirFileMode)
	if err != nil {
		return "", nil, nil, apperrors.Wrap(err)
	}

	tmpFileName = filepath.Join(data.TempDir, "sys-backup")
	tmpFile, err := os.Create(tmpFileName)
	if err != nil {
		return "", nil, nil, apperrors.Wrap(err)
	}

	var w io.Writer
	var encW, gzW io.WriteCloser
	var zstdW *zstd.Encoder
	w = tmpFile

	switch data.SysBackupSettings.Encryption.Format {
	case base.FileEncryptionFormatAge:
		encSecret := data.SysBackupSettings.Encryption.Secret.MustGetPlain()
		if encSecret == "" {
			return "", nil, nil, apperrors.NewMissing("Encryption secret")
		}
		recipient, err := age.NewScryptRecipient(encSecret)
		if err != nil {
			return "", nil, nil, apperrors.Wrap(err)
		}
		encW, err = age.Encrypt(w, recipient)
		if err != nil {
			return "", nil, nil, apperrors.Wrap(err)
		}
		w = encW
	case base.FileEncryptionNone: // Do nothing
	default:
		return "", nil, nil, apperrors.NewUnsupported(
			fmt.Sprintf("Encryption format '%v'", data.SysBackupSettings.Encryption.Format))
	}

	switch data.SysBackupSettings.Compression.Format {
	case base.FileCompressionFormatGzip:
		gzW = gzip.NewWriter(w)
		w = gzW
	case base.FileCompressionFormatZstd:
		zstdW, err = zstd.NewWriter(w)
		if err != nil {
			return "", nil, nil, apperrors.Wrap(err)
		}
		w = zstdW
	case base.FileCompressionNone: // Do nothing
	default:
		return "", nil, nil, apperrors.NewUnsupported(
			fmt.Sprintf("Compression format '%v'", data.SysBackupSettings.Compression.Format))
	}

	tarW = tar.NewWriter(w)

	closer = func() (err error) {
		if tarW != nil {
			if e := tarW.Close(); e != nil {
				err = errors.Join(err, e)
			}
		}
		if gzW != nil {
			if e := gzW.Close(); e != nil {
				err = errors.Join(err, e)
			}
		}
		if zstdW != nil {
			if e := zstdW.Close(); e != nil {
				err = errors.Join(err, e)
			}
		}
		if encW != nil {
			if e := encW.Close(); e != nil {
				err = errors.Join(err, e)
			}
		}
		if e := tmpFile.Close(); e != nil {
			err = errors.Join(err, e)
		}
		return err
	}

	return tmpFileName, tarW, closer, nil
}
