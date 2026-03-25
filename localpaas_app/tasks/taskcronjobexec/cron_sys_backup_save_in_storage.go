package taskcronjobexec

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/tiendc/gofn"

	"github.com/localpaas/localpaas/localpaas_app/apperrors"
	"github.com/localpaas/localpaas/localpaas_app/entity"
	"github.com/localpaas/localpaas/localpaas_app/pkg/applog"
	"github.com/localpaas/localpaas/localpaas_app/pkg/timeutil"
	"github.com/localpaas/localpaas/services/aws/s3"
)

func (e *Executor) sysDBBackupSaveResultInStorage(
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

	var s3Client *s3.Client
	var storageName string
	switch {
	case storage.S3 != nil:
		storageName = "AWS S3"
		s3Client, err = s3.NewClient(ctx, &s3.Config{
			AccessKeyID:     provider.AWS.AccessKeyID,
			SecretAccessKey: provider.AWS.SecretKey.MustGetPlain(),
			Endpoint:        storage.S3.Endpoint,
			Region:          gofn.Coalesce(storage.S3.Region, provider.AWS.Region),
			Bucket:          storage.S3.Bucket,
		})
	default:
		err = apperrors.NewUnsupported("Storage type")
	}
	if err != nil {
		return apperrors.Wrap(err)
	}

	backupFilePath := filepath.Join(data.BackupFileDir, data.BackupFileName)
	targetFilePath := filepath.Join(sysBackup.DestinationStorageDir, data.BackupFileName)

	backupFile, err := os.Open(backupFilePath)
	if err != nil {
		return apperrors.Wrap(err)
	}
	defer backupFile.Close()

	start := timeutil.NowUTC()
	_ = data.LogStore.Add(ctx, applog.NewOutFrame(fmt.Sprintf(
		"Start uploading file '%v' to '%v' bucket '%v'...",
		data.BackupFileName, storageName, storage.S3.Bucket), applog.TsNow))

	switch {
	case storage.S3 != nil:
		err = s3Client.UploadEx(ctx, storage.S3.Bucket, targetFilePath, 0, 0, backupFile)
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

	return nil
}
