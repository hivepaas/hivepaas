package taskcronjobexec

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
	"github.com/localpaas/localpaas/localpaas_app/pkg/applog"
	"github.com/localpaas/localpaas/localpaas_app/pkg/ulid"
)

const (
	sysBackupFilePrefix = "localpaas_backup_"
)

func (e *Executor) sysBackupSaveResultInLocal(
	ctx context.Context,
	sysBackup *entity.SystemBackup,
	tmpFile *os.File,
	data *sysBackupTaskData,
) (err error) {
	data.OutFileName = fmt.Sprintf("%s%s.jsonl", sysBackupFilePrefix,
		data.TimeNow.Truncate(time.Second).Format(time.RFC3339))
	if sysBackup.Compression {
		data.OutFileName += ".gz"
	}
	if sysBackup.EncryptionSecret.MustGetPlain() != "" {
		data.OutFileName += ".age"
	}

	data.OutFilePath = filepath.Join(data.BackupSaveDir, data.OutFileName)
	err = os.Rename(tmpFile.Name(), data.OutFilePath)
	if err != nil {
		_ = data.LogStore.Add(ctx, applog.NewErrFrame(
			"Failed to save backup data in file with error: "+err.Error(), applog.TsNow))
		return apperrors.Wrap(err)
	}

	// Save file details in to DB
	localFileSetting := &entity.Setting{
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
	localFile := &entity.File{
		FileKind:    base.FileKindSystemBackup,
		StorageType: base.FileStorageLocal,
		Mimetype:    "application/octet-stream",
		Path:        strings.TrimPrefix(data.BackupSaveDir, config.Current.AppPath),
	}
	localFileInfo, err := os.Stat(data.OutFilePath)
	if err != nil {
		return apperrors.Wrap(err)
	}
	localFile.Size = localFileInfo.Size()

	localFileSetting.MustSetData(localFile)
	data.LocalOutFile = localFileSetting

	_ = data.LogStore.Add(ctx, applog.NewOutFrame("Backup data saved into file: "+data.OutFileName,
		applog.TsNow))
	return nil
}
