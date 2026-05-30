package sysbackupserviceimpl

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/tiendc/gofn"

	"github.com/localpaas/localpaas/localpaas_app/apperrors"
	"github.com/localpaas/localpaas/localpaas_app/base"
	"github.com/localpaas/localpaas/localpaas_app/config"
	"github.com/localpaas/localpaas/localpaas_app/entity"
	"github.com/localpaas/localpaas/localpaas_app/infra/database"
	"github.com/localpaas/localpaas/localpaas_app/pkg/tasklog"
	"github.com/localpaas/localpaas/localpaas_app/pkg/ulid"
)

const (
	sysBackupFilePrefix = "localpaas_backup_"
)

func (s *service) sysBackupSaveResultInLocal(
	ctx context.Context,
	db database.IDB,
	bakTmpFile string,
	data *sysBackupData,
) (err error) {
	data.OutFileName = fmt.Sprintf("%s%s.tar", sysBackupFilePrefix,
		data.TimeNow.Truncate(time.Second).Format(time.RFC3339))

	switch data.SysBackupSettings.Compression.Format {
	case base.FileCompressionFormatGzip:
		data.OutFileName += ".gz"
	case base.FileCompressionFormatZstd:
		data.OutFileName += ".zst"
	case base.FileCompressionNone: // Do nothing
	default:
		return apperrors.NewUnsupported(
			fmt.Sprintf("Compression format '%v'", data.SysBackupSettings.Compression.Format))
	}

	switch data.SysBackupSettings.Encryption.Format {
	case base.FileEncryptionFormatAge:
		data.OutFileName += ".age"
	case base.FileEncryptionNone: // Do nothing
	default:
		return apperrors.NewUnsupported(
			fmt.Sprintf("Encryption format '%v'", data.SysBackupSettings.Encryption.Format))
	}

	data.OutFilePath = filepath.Join(data.BackupSaveDir, data.OutFileName)
	err = os.Rename(bakTmpFile, data.OutFilePath)
	if err != nil {
		_ = data.LogStore.Add(ctx, tasklog.NewErrFrame(
			"Failed to save backup data in file with error: "+err.Error(), tasklog.TsNow))
		return apperrors.Wrap(err)
	}

	// Save file details in to DB
	localFile := &entity.File{
		ID:          gofn.Must(ulid.NewStringULID()),
		Scope:       base.ObjectScopeGlobal,
		Type:        base.FileTypeSystemBackup,
		StorageType: base.FileStorageLocal,
		Status:      base.FileStatusActive,
		Name:        data.OutFileName,
		Mimetype:    "application/octet-stream",
		Path:        strings.TrimPrefix(data.BackupSaveDir, config.Current.AppPath),
		CreatedAt:   data.TimeNow,
		UpdatedAt:   data.TimeNow,
	}
	localFileInfo, err := os.Stat(data.OutFilePath)
	if err != nil {
		return apperrors.Wrap(err)
	}
	localFile.Size = localFileInfo.Size()
	data.LocalOutFile = localFile

	err = s.fileRepo.Insert(ctx, db, localFile)
	if err != nil {
		_ = data.LogStore.Add(ctx, tasklog.NewOutFrame("Failed to save file into DB with error: "+
			err.Error(), tasklog.TsNow))
		return apperrors.Wrap(err)
	}

	_ = data.LogStore.Add(ctx, tasklog.NewOutFrame("Backup data saved into file: "+data.OutFileName,
		tasklog.TsNow))
	return nil
}
