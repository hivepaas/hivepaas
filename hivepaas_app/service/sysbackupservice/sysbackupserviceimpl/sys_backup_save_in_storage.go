package sysbackupserviceimpl

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"time"

	"github.com/tiendc/gofn"

	"github.com/hivepaas/hivepaas/hivepaas_app/apperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/infra/database"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/tasklog"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/timeutil"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/ulid"
	"github.com/hivepaas/hivepaas/services/aws/s3"
)

func (s *service) sysBackupSaveResultInStorage(
	ctx context.Context,
	db database.IDB,
	data *sysBackupData,
) (err error) {
	if data.SysBackupSettings.CloudStorage.ID == "" {
		return nil
	}
	storageSetting := data.RefObjects.RefSettings[data.SysBackupSettings.CloudStorage.ID]
	if storageSetting == nil {
		return nil
	}

	var storageName string
	var storageBucket string
	var uploadFunc func(targetFilePath string, data io.Reader) error

	switch base.CloudStorageKind(storageSetting.Kind) {
	case base.CloudStorageKindS3:
		s3Client, err := s3.NewClientFromSetting(ctx, storageSetting)
		if err != nil {
			return apperrors.Wrap(err)
		}
		storageName = "AWS S3"
		storageBucket = s3Client.Config.Bucket
		uploadFunc = func(targetFilePath string, input io.Reader) error {
			return s3Client.UploadEx(ctx, storageBucket, targetFilePath,
				0, 0, input)
		}
	default:
		return apperrors.Wrap(apperrors.ErrStorageTypeUnsupported).WithParam("Type", storageSetting.Kind)
	}

	targetFilePath := filepath.Join(data.SysBackupSettings.CloudStorage.DestinationDir, data.OutFileName)
	backupFile, err := os.Open(data.OutFilePath)
	if err != nil {
		return apperrors.Wrap(err)
	}
	defer backupFile.Close()

	start := timeutil.NowUTC()
	_ = data.LogStore.Add(ctx, tasklog.NewOutFrame(fmt.Sprintf(
		"Start uploading file '%v' to '%v' bucket '%v'...",
		data.OutFileName, storageName, storageBucket), tasklog.TsNow))

	err = uploadFunc(targetFilePath, backupFile)
	if err != nil {
		_ = data.LogStore.Add(ctx, tasklog.NewWarnFrame(
			"Failed to upload backup file to "+storageName+" with error: "+err.Error(), tasklog.TsNow))
		return apperrors.Wrap(err)
	}
	_ = data.LogStore.Add(ctx, tasklog.NewOutFrame("Backup file uploaded to "+storageName+
		" in "+time.Since(start).String(), tasklog.TsNow))

	localFile := data.LocalOutFile
	remoteFile := &entity.File{
		ID:          gofn.Must(ulid.NewStringULID()),
		Scope:       base.ObjectScopeGlobal,
		Type:        base.FileTypeSystemBackup,
		Status:      base.FileStatusActive,
		StorageType: base.FileStorageCloud,
		StorageID:   data.SysBackupSettings.CloudStorage.ID,
		Bucket:      storageBucket,
		Name:        data.OutFileName,
		Path:        targetFilePath,
		Mimetype:    localFile.Mimetype,
		Size:        localFile.Size,
		CreatedAt:   data.TimeNow,
		UpdatedAt:   data.TimeNow,
	}
	data.RemoteOutFile = remoteFile

	err = s.fileRepo.Insert(ctx, db, remoteFile)
	if err != nil {
		_ = data.LogStore.Add(ctx, tasklog.NewOutFrame("Failed to save file into DB with error: "+
			err.Error(), tasklog.TsNow))
		return apperrors.Wrap(err)
	}

	return nil
}
