package systembackupdto

import (
	"time"

	vld "github.com/tiendc/go-validator"

	"github.com/localpaas/localpaas/localpaas_app/apperrors"
	"github.com/localpaas/localpaas/localpaas_app/base"
	"github.com/localpaas/localpaas/localpaas_app/basedto"
	"github.com/localpaas/localpaas/localpaas_app/entity"
	"github.com/localpaas/localpaas/localpaas_app/pkg/timeutil"
	"github.com/localpaas/localpaas/localpaas_app/usecase/settings"
)

type UpdateSystemBackupReq struct {
	settings.UpdateSettingReq
	*SystemBackupBaseReq
}

type SystemBackupBaseReq struct {
	Status                base.SettingStatus                `json:"status"`
	ScheduleInterval      timeutil.Duration                 `json:"scheduleInterval"`
	ScheduleFrom          time.Time                         `json:"scheduleFrom"`
	DBBackupConfig        *DBBackupConfigReq                `json:"dbBackupConfig"`
	Compression           bool                              `json:"compression"`
	EncryptionSecret      string                            `json:"encryptionSecret"`
	DestinationStorage    basedto.ObjectIDReq               `json:"destinationStorage"`
	DestinationStorageDir string                            `json:"destinationStorageDir"`
	LocalBackupRetention  timeutil.Duration                 `json:"localBackupRetention"`
	Notification          *basedto.BaseEventNotificationReq `json:"notification"`
}

func (req *SystemBackupBaseReq) ToEntity() *entity.SystemBackup {
	return &entity.SystemBackup{
		ScheduleInterval:      req.ScheduleInterval,
		ScheduleFrom:          req.ScheduleFrom,
		DBBackupConfig:        req.DBBackupConfig.ToEntity(),
		Compression:           req.Compression,
		EncryptionSecret:      entity.NewEncryptedField(req.EncryptionSecret),
		DestinationStorage:    entity.ObjectID{ID: req.DestinationStorage.ID},
		DestinationStorageDir: req.DestinationStorageDir,
		LocalBackupRetention:  req.LocalBackupRetention,
		Notification:          req.Notification.ToEntity(),
	}
}

type DBBackupConfigReq struct {
	BackupDeletedObjects bool `json:"backupDeletedObjects"`
}

func (req *DBBackupConfigReq) ToEntity() *entity.DBBackupConfig {
	if req == nil {
		return nil
	}
	return &entity.DBBackupConfig{
		BackupDeletedObjects: req.BackupDeletedObjects,
	}
}

func (req *SystemBackupBaseReq) validate(_ string) []vld.Validator {
	// TODO: add validation
	return nil
}

func NewUpdateSystemBackupReq() *UpdateSystemBackupReq {
	return &UpdateSystemBackupReq{}
}

func (req *UpdateSystemBackupReq) ModifyRequest() error {
	if !req.ScheduleFrom.IsZero() {
		req.ScheduleFrom = req.ScheduleFrom.Truncate(time.Minute)
	}
	return nil
}

// Validate implements interface basedto.ReqValidator
func (req *UpdateSystemBackupReq) Validate() apperrors.ValidationErrors {
	var validators []vld.Validator
	validators = append(validators, req.validate("")...)
	return apperrors.NewValidationErrors(vld.Validate(validators...))
}

type UpdateSystemBackupResp struct {
	Meta *basedto.Meta `json:"meta"`
}
