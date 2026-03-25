package systembackupdto

import (
	"time"

	vld "github.com/tiendc/go-validator"

	"github.com/localpaas/localpaas/localpaas_app/apperrors"
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
	ScheduleInterval      timeutil.Duration                  `json:"scheduleInterval"`
	ScheduleFrom          time.Time                          `json:"scheduleFrom"`
	DBBackupConfig        *DBBackupConfigResp                `json:"dbBackupConfig"`
	Compression           bool                               `json:"compression"`
	EncryptionSecret      string                             `json:"encryptionSecret"`
	DestinationStorage    *settings.BaseSettingResp          `json:"destinationStorage"`
	DestinationStorageDir string                             `json:"destinationStorageDir"`
	LocalBackupRetention  timeutil.Duration                  `json:"localBackupRetention"`
	Notification          *basedto.BaseEventNotificationResp `json:"notification"`
}

type DBBackupConfigResp struct {
	BackupDeletedObjects bool `json:"backupDeletedObjects"`
}

func (resp *SystemBackupResp) CopyEncryptionSecret(field entity.EncryptedField) error {
	resp.EncryptionSecret = field.String()
	return nil
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

	if config.DestinationStorage.ID != "" {
		setting := refObjects.RefSettings[config.DestinationStorage.ID]
		resp.DestinationStorage, _ = settings.TransformSettingBase(setting)
	} else {
		resp.DestinationStorage = nil
	}

	resp.Notification = basedto.TransformBaseEventNotification(config.Notification, refObjects)
	return resp, nil
}
