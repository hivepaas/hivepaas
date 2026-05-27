package systembackupdto

import (
	"time"

	vld "github.com/tiendc/go-validator"

	"github.com/localpaas/localpaas/localpaas_app/apperrors"
	"github.com/localpaas/localpaas/localpaas_app/base"
	"github.com/localpaas/localpaas/localpaas_app/basedto"
	"github.com/localpaas/localpaas/localpaas_app/entity"
	"github.com/localpaas/localpaas/localpaas_app/pkg/copier"
	"github.com/localpaas/localpaas/localpaas_app/pkg/timeutil"
	"github.com/localpaas/localpaas/localpaas_app/usecase/settings"
)

type GetSystemBackupReq struct {
	settings.GetSettingReq
}

func NewGetSystemBackupReq() *GetSystemBackupReq {
	return &GetSystemBackupReq{}
}

func (req *GetSystemBackupReq) Validate() apperrors.ValidationErrors {
	var validators []vld.Validator
	validators = append(validators, req.GetSettingReq.Validate()...)
	return apperrors.NewValidationErrors(vld.Validate(validators...))
}

type GetSystemBackupResp struct {
	Meta *basedto.Meta     `json:"meta"`
	Data *SystemBackupResp `json:"data"`
}

type SystemBackupResp struct {
	*settings.BaseSettingResp
	ScheduleInterval timeutil.Duration                  `json:"scheduleInterval"`
	ScheduleCronExpr string                             `json:"scheduleCronExpr"`
	ScheduleFrom     time.Time                          `json:"scheduleFrom"`
	Compression      *SystemBackupCompressionResp       `json:"compression"`
	Encryption       *SystemBackupEncryptionResp        `json:"encryption"`
	CloudStorage     *SystemBackupCloudStorageResp      `json:"cloudStorage"`
	DBBackupConfig   *SystemBackupDBConfigResp          `json:"dbBackupConfig"`
	Notification     *basedto.BaseEventNotificationResp `json:"notification"`
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
	config := setting.MustAsSystemBackup().MustDecrypt()
	if err = copier.Copy(&resp, config); err != nil {
		return nil, apperrors.Wrap(err)
	}

	resp.BaseSettingResp, err = settings.TransformSettingBase(setting)
	if err != nil {
		return nil, apperrors.Wrap(err)
	}

	if config.CloudStorage.ID != "" {
		setting := refObjects.RefSettings[config.CloudStorage.ID]
		resp.CloudStorage.BaseSettingResp, _ = settings.TransformSettingBase(setting)
	} else {
		resp.CloudStorage = nil
	}

	resp.Notification = basedto.TransformBaseEventNotification(config.Notification, refObjects)
	return resp, nil
}
