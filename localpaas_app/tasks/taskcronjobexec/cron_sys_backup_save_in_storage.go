package taskcronjobexec

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/tiendc/gofn"

	"github.com/localpaas/localpaas/localpaas_app/apperrors"
	"github.com/localpaas/localpaas/localpaas_app/base"
	"github.com/localpaas/localpaas/localpaas_app/entity"
	"github.com/localpaas/localpaas/localpaas_app/pkg/applog"
	"github.com/localpaas/localpaas/localpaas_app/pkg/timeutil"
	"github.com/localpaas/localpaas/localpaas_app/pkg/ulid"
	"github.com/localpaas/localpaas/services/aws/s3"
)

func (e *Executor) sysBackupSaveResultInStorage(
	ctx context.Context,
	sysBackup *entity.SystemBackup,
	data *sysBackupTaskData,
) (err error) {
	if sysBackup.DestinationStorage.ID == "" {
		return nil
	}
	storageSttg := data.RefObjects.RefSettings[sysBackup.DestinationStorage.ID]
	if storageSttg == nil {
		return nil
	}
	storage := storageSttg.MustAsCloudStorage()
	providerSttg := data.RefObjects.RefSettings[storage.Provider.ID]
	if providerSttg == nil {
		return nil
	}
	provider := providerSttg.MustAsCloudProvider()
	storage.RefProvider = provider

	var s3Client *s3.Client
	var storageName string
	var storageBucket string
	switch {
	case storage.S3 != nil:
		storageName = "AWS S3"
		storageBucket = storage.S3.Bucket
		s3Client, err = s3.NewClientFromStorage(ctx, storage)
	default:
		err = apperrors.NewUnsupported("Storage type")
	}
	if err != nil {
		return apperrors.Wrap(err)
	}

	targetFilePath := filepath.Join(sysBackup.DestinationStorageDir, data.OutFileName)
	backupFile, err := os.Open(data.OutFilePath)
	if err != nil {
		return apperrors.Wrap(err)
	}
	defer backupFile.Close()

	start := timeutil.NowUTC()
	_ = data.LogStore.Add(ctx, applog.NewOutFrame(fmt.Sprintf(
		"Start uploading file '%v' to '%v' bucket '%v'...",
		data.OutFileName, storageName, storage.S3.Bucket), applog.TsNow))

	switch {
	case storage.S3 != nil:
		err = s3Client.UploadEx(ctx, storageBucket, targetFilePath, 0, 0, backupFile)
	default:
		err = apperrors.NewUnsupported("Storage type")
	}

	if err != nil {
		_ = data.LogStore.Add(ctx, applog.NewWarnFrame(
			"Failed to upload backup file to "+storageName+" with error: "+err.Error(), applog.TsNow))
		return apperrors.Wrap(err)
	}
	_ = data.LogStore.Add(ctx, applog.NewOutFrame("Backup file uploaded to "+storageName+
		" in "+time.Since(start).String(), applog.TsNow))

	localFile := data.LocalOutFile.MustAsFile()
	remoteFileSetting := &entity.Setting{
		ID:        gofn.Must(ulid.NewStringULID()),
		Scope:     base.SettingScopeGlobal,
		Type:      base.SettingTypeFile,
		Kind:      string(base.FileKindSystemBackup),
		Status:    base.SettingStatusActive,
		Name:      data.OutFileName,
		Version:   entity.CurrentFileVersion,
		CreatedAt: data.TimeNow,
		UpdatedAt: data.TimeNow,
	}

	remoteFile := &entity.File{
		FileKind:    base.FileKindSystemBackup,
		StorageType: base.FileStorageCloud,
		Storage:     entity.ObjectID{ID: sysBackup.DestinationStorage.ID},
		Bucket:      storageBucket,
		Mimetype:    localFile.Mimetype,
		Size:        localFile.Size,
		Path:        sysBackup.DestinationStorageDir,
	}

	remoteFileSetting.MustSetData(remoteFile)
	data.RemoteOutFile = remoteFileSetting

	return nil
}
