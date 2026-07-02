package systembackupdto

import (
	"time"

	vld "github.com/tiendc/go-validator"

	"github.com/hivepaas/hivepaas/hivepaas_app/apperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/basedto"
	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/timeutil"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/settings"
)

type UpdateSystemBackupReq struct {
	settings.UpdateSettingReq
	*SystemBackupBaseReq
}

type SystemBackupBaseReq struct {
	Status         base.SettingStatus                `json:"status"`
	Schedule       ScheduleReq                       `json:"schedule"`
	Compression    SystemBackupCompressionReq        `json:"compression"`
	Encryption     SystemBackupEncryptionReq         `json:"encryption"`
	CloudStorage   SystemBackupCloudStorageReq       `json:"cloudStorage"`
	DBBackupConfig SystemBackupDBConfigReq           `json:"dbBackupConfig"`
	Notification   *basedto.BaseEventNotificationReq `json:"notification"`
}

func (req *SystemBackupBaseReq) ToEntity() *entity.SystemBackup {
	return &entity.SystemBackup{
		Schedule:       req.Schedule.ToEntity(),
		Compression:    req.Compression.ToEntity(),
		Encryption:     req.Encryption.ToEntity(),
		CloudStorage:   req.CloudStorage.ToEntity(),
		DBBackupConfig: req.DBBackupConfig.ToEntity(),
		Notification:   req.Notification.ToEntity(),
	}
}

type ScheduleReq struct {
	CronExpr    string            `json:"cronExpr"` // cronExpr and interval are mutually exclusive
	Interval    timeutil.Duration `json:"interval"`
	InitialTime time.Time         `json:"initialTime"`
}

func (req *ScheduleReq) ToEntity() entity.SchedJobSchedule {
	return entity.SchedJobSchedule{
		CronExpr:    req.CronExpr,
		Interval:    req.Interval,
		InitialTime: req.InitialTime,
	}
}

type SystemBackupCompressionReq struct {
	Format base.FileCompressionFormat `json:"format,omitempty"`
}

func (req *SystemBackupCompressionReq) ToEntity() entity.SystemBackupCompression {
	return entity.SystemBackupCompression{
		Format: req.Format,
	}
}

func (req *SystemBackupCompressionReq) validate(field string) (res []vld.Validator) {
	if field != "" {
		field += "."
	}
	res = append(res, basedto.ValidateStrIn(&req.Format, false,
		base.AllFileCompressionFormats, field+"format")...)
	return res
}

type SystemBackupEncryptionReq struct {
	Format base.FileEncryptionFormat `json:"format,omitempty"`
	Secret string                    `json:"secret,omitzero"`
}

func (req *SystemBackupEncryptionReq) ToEntity() entity.SystemBackupEncryption {
	return entity.SystemBackupEncryption{
		Format: req.Format,
		Secret: entity.NewEncryptedField(req.Secret),
	}
}

func (req *SystemBackupEncryptionReq) validate(field string) (res []vld.Validator) {
	if field != "" {
		field += "."
	}
	res = append(res, basedto.ValidateStrIn(&req.Format, false,
		base.AllFileEncryptionFormats, field+"format")...)
	return res
}

type SystemBackupCloudStorageReq struct {
	ID             string `json:"id"`
	DestinationDir string `json:"destinationDir"`
}

func (req *SystemBackupCloudStorageReq) ToEntity() entity.SystemBackupCloudStorage {
	return entity.SystemBackupCloudStorage{
		ID:             req.ID,
		DestinationDir: req.DestinationDir,
	}
}

func (req *SystemBackupCloudStorageReq) validate(field string) (res []vld.Validator) {
	if field != "" {
		field += "."
	}
	res = append(res, basedto.ValidateID(&req.ID, false, field+"id")...)
	// TODO: add validation
	return res
}

type SystemBackupDBConfigReq struct {
	BackupDeletedObjects bool `json:"backupDeletedObjects"`
}

func (req *SystemBackupDBConfigReq) ToEntity() entity.SystemBackupDBConfig {
	return entity.SystemBackupDBConfig{
		BackupDeletedObjects: req.BackupDeletedObjects,
	}
}

func (req *SystemBackupBaseReq) validate(field string) (res []vld.Validator) {
	if field != "" {
		field += "."
	}
	sched := req.Schedule.ToEntity()
	res = append(res, vld.Must((&sched).IsValid() == nil).OnError(
		vld.SetField(field+"schedule.Interval|schedule.CronExpr", nil),
		vld.SetCustomKey("ERR_VLD_VALUE_REQUIRED_ONLY"),
	))
	res = append(res, basedto.ValidateTime(&req.Schedule.InitialTime, true,
		time.Now().Add(-timeutil.Dur365Days), time.Time{}, field+"schedule.initialTime")...)
	res = append(res, req.Compression.validate(field+"compression")...)
	res = append(res, req.Encryption.validate(field+"encryption")...)
	res = append(res, req.CloudStorage.validate(field+"cloudStorage")...)
	res = append(res, req.Notification.Validate(field+"notification")...)
	return res
}

func NewUpdateSystemBackupReq() *UpdateSystemBackupReq {
	return &UpdateSystemBackupReq{}
}

func (req *UpdateSystemBackupReq) ModifyRequest() error {
	if req.Schedule.InitialTime.IsZero() {
		req.Schedule.InitialTime = timeutil.NowUTC()
	}
	req.Schedule.InitialTime = req.Schedule.InitialTime.Truncate(time.Minute)
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
