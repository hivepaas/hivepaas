package systembackupdto

import (
	"time"

	vld "github.com/tiendc/go-validator"

	"github.com/hivepaas/hivepaas/hivepaas_app/apperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/basedto"
	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/copier"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/timeutil"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/settings"
)

type GetSystemBackupReq struct {
	settings.GetUniqueSettingReq
}

func NewGetSystemBackupReq() *GetSystemBackupReq {
	return &GetSystemBackupReq{}
}

func (req *GetSystemBackupReq) Validate() apperrors.ValidationErrors {
	var validators []vld.Validator
	validators = append(validators, req.GetUniqueSettingReq.Validate()...)
	return apperrors.NewValidationErrors(vld.Validate(validators...))
}

type GetSystemBackupResp struct {
	Meta *basedto.Meta     `json:"meta"`
	Data *SystemBackupResp `json:"data"`
}

type SystemBackupResp struct {
	*settings.BaseSettingResp
	Schedule       *ScheduleResp                      `json:"schedule"`
	Compression    *SystemBackupCompressionResp       `json:"compression"`
	Encryption     *SystemBackupEncryptionResp        `json:"encryption"`
	CloudStorage   *SystemBackupCloudStorageResp      `json:"cloudStorage"`
	DBBackupConfig *SystemBackupDBConfigResp          `json:"dbBackupConfig"`
	Notification   *basedto.BaseEventNotificationResp `json:"notification"`

	// Calculated fields
	NextRuns []time.Time `json:"nextRuns"`
}

type ScheduleResp struct {
	CronExpr    string            `json:"cronExpr,omitempty"` // cronExpr and interval are mutually exclusive
	Interval    timeutil.Duration `json:"interval,omitempty"`
	InitialTime time.Time         `json:"initialTime"`
}

type SystemBackupCompressionResp struct {
	Format base.FileCompressionFormat `json:"format,omitempty"`
}

type SystemBackupEncryptionResp struct {
	Format base.FileEncryptionFormat `json:"format,omitempty"`
	Secret string                    `json:"secret,omitzero"`
}

func (resp *SystemBackupEncryptionResp) CopySecret(field entity.EncryptedField) error {
	resp.Secret = field.String()
	return nil
}

type SystemBackupCloudStorageResp struct {
	*settings.BaseSettingResp
	DestinationDir string `json:"destinationDir,omitempty"`
}

type SystemBackupDBConfigResp struct {
	BackupDeletedObjects bool `json:"backupDeletedObjects"`
}

func TransformSystemBackup(
	setting *entity.Setting,
	refObjects *entity.RefObjects,
) (resp *SystemBackupResp, err error) {
	config := setting.MustAsSystemBackup()
	err = config.Decrypt()
	if err != nil {
		return nil, apperrors.New(err)
	}
	if err = copier.Copy(&resp, config); err != nil {
		return nil, apperrors.New(err)
	}

	resp.BaseSettingResp, err = settings.TransformSettingBase(setting)
	if err != nil {
		return nil, apperrors.New(err)
	}

	if config.CloudStorage.ID != "" {
		setting := refObjects.RefSettings[config.CloudStorage.ID]
		resp.CloudStorage.BaseSettingResp, _ = settings.TransformSettingBase(setting)
	} else {
		resp.CloudStorage = nil
	}

	resp.Notification = basedto.TransformBaseEventNotification(config.Notification, refObjects)

	// Add next runs
	resp.NextRuns, _ = config.Schedule.CalcNextRuns(time.Now(), 5) //nolint

	return resp, nil
}
