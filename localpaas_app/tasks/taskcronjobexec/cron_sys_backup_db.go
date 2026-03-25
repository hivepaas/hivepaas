package taskcronjobexec

import (
	"compress/gzip"
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"reflect"
	"time"

	"filippo.io/age"
	"github.com/tiendc/gofn"

	"github.com/localpaas/localpaas/localpaas_app/apperrors"
	"github.com/localpaas/localpaas/localpaas_app/config"
	"github.com/localpaas/localpaas/localpaas_app/entity"
	"github.com/localpaas/localpaas/localpaas_app/infra/database"
	"github.com/localpaas/localpaas/localpaas_app/pkg/applog"
	"github.com/localpaas/localpaas/localpaas_app/pkg/jsonl"
	"github.com/localpaas/localpaas/localpaas_app/pkg/timeutil"
)

const (
	// 0755 grants read/write/execute for owner, read/execute for group/others
	// 0644 grants read/write for owner, read-only for group/others
	backupDirFileMode = 0o755

	backupFilePrefix = "localpaas_backup_"
)

func (e *Executor) sysDBBackup(
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

	tmpFile, jsonlW, closer, err := e.sysDBBackupCreateWriter(sysBackup, data)
	if err != nil {
		return apperrors.Wrap(err)
	}
	defer func() {
		if closer != nil {
			_ = closer(true)
		}
	}()

	// Start the data backup
	err = e.sysDBBackupProcess(ctx, db, sysBackup, jsonlW, data)
	if err != nil {
		return apperrors.Wrap(err)
	}

	_ = closer(false) // Flush data in writers, but not remove the temp file
	closer = nil

	// Save the result in a file
	err = e.sysDBBackupSaveResultInFile(ctx, sysBackup, tmpFile, data)
	if err != nil {
		return apperrors.Wrap(err)
	}

	// Upload backup file to cloud storage if configured
	err = e.sysDBBackupSaveResultInStorage(ctx, sysBackup, data)
	if err != nil {
		return apperrors.Wrap(err)
	}

	// Remove outdated backup files
	err = e.sysDBBackupRemoveOldFiles(ctx, sysBackup, data)
	if err != nil {
		return apperrors.Wrap(err)
	}

	return nil
}

func (e *Executor) sysDBBackupCreateWriter(
	sysBackup *entity.SystemBackup,
	data *sysBackupTaskData,
) (tmpFile *os.File, jsonlW *jsonl.Writer, closer func(bool) error, err error) {
	// Make sure the backup directory exist
	data.BackupDir = config.Current.DataPathSystemBackup()
	data.BackupFileDir = config.Current.DataPathSystemBackupFiles()

	err = os.MkdirAll(data.BackupDir, backupDirFileMode)
	if err != nil {
		return nil, nil, nil, apperrors.Wrap(err)
	}

	tmpDir := filepath.Join(data.BackupDir, "tmp")
	err = os.MkdirAll(tmpDir, backupDirFileMode)
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

func (e *Executor) sysDBBackupProcess(
	ctx context.Context,
	db database.IDB,
	sysBackup *entity.SystemBackup,
	jsonlW *jsonl.Writer,
	data *sysBackupTaskData,
) (err error) {
	start := timeutil.NowUTC()
	_ = data.LogStore.Add(ctx, applog.NewWarnFrame("Start backing up data from DB...", applog.TsNow))

	defer func() {
		duration := timeutil.NowUTC().Sub(start)
		if err != nil {
			_ = data.LogStore.Add(ctx, applog.NewWarnFrame("Data backup finished in "+duration.String()+
				" with error: "+err.Error(), applog.TsNow))
		} else {
			_ = data.LogStore.Add(ctx, applog.NewOutFrame("Data backup finished in "+duration.String(),
				applog.TsNow))
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

	backupDeletedObjects := true
	if sysBackup.DBBackupConfig != nil {
		backupDeletedObjects = sysBackup.DBBackupConfig.BackupDeletedObjects
	}

	for _, model := range sysBackupDBModels {
		// TODO: use cursor to speed up the backup process
		offset := 0
		for {
			pageSize := gofn.Coalesce(model.PageSize, sysBackupSqlPageSize)
			q := db.NewSelect().Model(model.Model).Limit(pageSize).Offset(offset)
			if len(model.Orders) > 0 {
				q = q.Order(model.Orders...)
			}
			if backupDeletedObjects && !model.NoSoftDelete {
				q = q.WhereAllWithDeleted()
			}
			err = q.Scan(ctx)
			if err != nil {
				return apperrors.Wrap(err)
			}

			// Reflection to get the length of the slice
			val := reflect.ValueOf(model.Model).Elem()
			if val.Len() == 0 {
				break
			}

			err = jsonlW.WriteChunk(jsonl.NewChunk(model.Type, val.Interface()))
			if err != nil {
				return apperrors.Wrap(err)
			}

			if val.Len() < pageSize {
				break
			}
			offset += pageSize
		}
	}

	return nil
}

func (e *Executor) sysDBBackupSaveResultInFile(
	ctx context.Context,
	sysBackup *entity.SystemBackup,
	tmpFile *os.File,
	data *sysBackupTaskData,
) (err error) {
	data.BackupFileName = fmt.Sprintf("%s%s.jsonl", backupFilePrefix,
		data.TimeNow.Truncate(time.Second).Format(time.RFC3339))
	if sysBackup.Compression {
		data.BackupFileName += ".gz"
	}
	if sysBackup.EncryptionSecret.MustGetEncrypted() != "" {
		data.BackupFileName += ".age"
	}

	err = os.Rename(tmpFile.Name(), filepath.Join(data.BackupFileDir, data.BackupFileName))
	if err != nil {
		_ = data.LogStore.Add(ctx, applog.NewErrFrame(
			"Failed to save backup data in file with error: "+err.Error(), applog.TsNow))
		return apperrors.Wrap(err)
	}

	_ = data.LogStore.Add(ctx, applog.NewOutFrame("Backup data saved into file: "+data.BackupFileName,
		applog.TsNow))
	return nil
}
