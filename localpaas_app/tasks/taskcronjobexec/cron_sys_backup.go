package taskcronjobexec

import (
	"compress/gzip"
	"context"
	"errors"
	"io"
	"os"
	"path/filepath"
	"time"

	"filippo.io/age"

	"github.com/localpaas/localpaas/localpaas_app/apperrors"
	"github.com/localpaas/localpaas/localpaas_app/config"
	"github.com/localpaas/localpaas/localpaas_app/entity"
	"github.com/localpaas/localpaas/localpaas_app/infra/database"
	"github.com/localpaas/localpaas/localpaas_app/pkg/jsonl"
	"github.com/localpaas/localpaas/localpaas_app/pkg/timeutil"
)

const (
	// 0755 grants read/write/execute for owner, read/execute for group/others
	// 0644 grants read/write for owner, read-only for group/others
	sysBackupDirFileMode = 0o755
)

type sysBackupTaskData struct {
	*taskData
	TaskOutput    *entity.TaskSystemBackupOutput
	BackupRootDir string
	BackupSaveDir string
	OutFileName   string
	OutFilePath   string
	TimeNow       time.Time
	LocalOutFile  *entity.Setting
	RemoteOutFile *entity.Setting
}

func (e *Executor) cronExecSystemBackup(
	ctx context.Context,
	db database.IDB,
	data *taskData,
) error {
	setting := data.RefObjects.RefSettings[data.CronJob.TargetSetting.ID]
	if setting == nil {
		return apperrors.NewNotFound("System backup settings")
	}
	sysBackup := setting.MustAsSystemBackup()

	taskData := &sysBackupTaskData{
		taskData: data,
		TaskOutput: &entity.TaskSystemBackupOutput{
			DBBackup: &entity.DBBackupOutput{},
		},
		TimeNow: timeutil.NowUTC(),
	}

	// Backup DB
	err := e.sysBackup(ctx, db, sysBackup, taskData)

	// Assign back the result output
	data.Task.MustSetOutput(taskData.TaskOutput)

	return err
}

func (e *Executor) sysBackup(
	ctx context.Context,
	db database.IDB,
	sysBackup *entity.SystemBackup,
	data *sysBackupTaskData,
) (err error) {
	if sysBackup == nil {
		return nil
	}

	defer func() {
		if err != nil {
			data.TaskOutput.DBBackup.Error = err.Error()
		}
	}()

	tmpFile, jsonlW, closer, err := e.sysBackupCreateWriter(sysBackup, data)
	if err != nil {
		return apperrors.Wrap(err)
	}
	defer func() {
		if closer != nil {
			_ = closer(true)
		}
	}()

	// Write header to the backup file
	err = jsonlW.WriteMetadata(jsonl.Metadata{
		Type:      "system-backup",
		Timestamp: data.TimeNow.Truncate(time.Second),
	})
	if err != nil {
		return apperrors.Wrap(err)
	}

	// Start the data backup
	err = e.sysBackupDB(ctx, db, sysBackup, jsonlW, data)
	if err != nil {
		return apperrors.Wrap(err)
	}

	err = e.sysBackupFiles(ctx, db, jsonlW, data)
	if err != nil {
		return apperrors.Wrap(err)
	}

	_ = closer(false) // Flush data in writers, but not remove the temp file
	closer = nil

	// Save the result in a file
	err = e.sysBackupSaveResultInLocal(ctx, sysBackup, tmpFile, data)
	if err != nil {
		return apperrors.Wrap(err)
	}

	// Upload backup file to cloud storage if configured
	err = e.sysBackupSaveResultInStorage(ctx, sysBackup, data)
	if err != nil {
		return apperrors.Wrap(err)
	}

	// Remove outdated backup files
	err = e.sysBackupRemoveOldFiles(ctx, db, sysBackup, data)
	if err != nil {
		return apperrors.Wrap(err)
	}

	return nil
}

func (e *Executor) sysBackupCreateWriter(
	sysBackup *entity.SystemBackup,
	data *sysBackupTaskData,
) (tmpFile *os.File, jsonlW *jsonl.Writer, closer func(bool) error, err error) {
	// Make sure the backup directory exist
	data.BackupRootDir = config.Current.DataPathSystemBackup()
	data.BackupSaveDir = config.Current.DataPathSystemBackupFiles()

	err = os.MkdirAll(data.BackupRootDir, sysBackupDirFileMode)
	if err != nil {
		return nil, nil, nil, apperrors.Wrap(err)
	}

	tmpDir := filepath.Join(data.BackupRootDir, "tmp")
	err = os.MkdirAll(tmpDir, sysBackupDirFileMode)
	if err != nil {
		return nil, nil, nil, apperrors.Wrap(err)
	}

	tmpFile, err = os.CreateTemp(tmpDir, "*.bak")
	if err != nil {
		return nil, nil, nil, apperrors.Wrap(err)
	}

	defer func() {
		if err != nil {
			_ = os.Remove(tmpFile.Name())
		}
	}()

	var w io.Writer
	var encW, gzW io.WriteCloser
	w = tmpFile

	encSecret := sysBackup.EncryptionSecret.MustGetPlain()
	if encSecret != "" {
		recipient, err := age.NewScryptRecipient(encSecret)
		if err != nil {
			return nil, nil, nil, apperrors.Wrap(err)
		}
		encW, err = age.Encrypt(w, recipient)
		if err != nil {
			return nil, nil, nil, apperrors.Wrap(err)
		}
		w = encW
	}
	if sysBackup.Compression {
		gzW = gzip.NewWriter(w)
		w = gzW
	}

	jsonlW = jsonl.NewWriter(w)

	closer = func(removeTmpFile bool) (err error) {
		if e := jsonlW.Close(); e != nil {
			err = errors.Join(err, e)
		}
		if gzW != nil {
			if e := gzW.Close(); e != nil {
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
		if removeTmpFile {
			_ = os.Remove(tmpFile.Name()) // Ignore this error as the temp file may not exist
		}
		return err
	}

	return tmpFile, jsonlW, closer, nil
}
