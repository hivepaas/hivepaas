package systembackupdto

import (
	"time"

	vld "github.com/tiendc/go-validator"

	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/basedto"
	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/timeutil"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/specservice/specmodel"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/settings"
)

// specPassphraseMaxLen is as long as a spec export's passphrase may be.
const specPassphraseMaxLen = 256

type UpdateSystemBackupReq struct {
	settings.UpdateUniqueSettingReq
	*SystemBackupBaseReq
}

type SystemBackupBaseReq struct {
	Status   base.SettingStatus `json:"status"`
	Schedule ScheduleReq        `json:"schedule"`
	// What goes in: one at least.
	IncludeDB   bool `json:"includeDB"`
	IncludeSpec bool `json:"includeSpec"`
	// SpecSecrets is how the spec holds secrets; SpecPassphrase is required
	// when they are encrypted.
	SpecSecrets    specmodel.SecretsMode `json:"specSecrets"`
	SpecPassphrase string                `json:"specPassphrase"`
	// TargetRepository is a backup repository at the global scope.
	TargetRepository basedto.ObjectIDReq               `json:"targetRepository"`
	Notification     *basedto.BaseEventNotificationReq `json:"notification"`
}

func (req *SystemBackupBaseReq) ToEntity() *entity.SystemBackup {
	backup := &entity.SystemBackup{
		Schedule:         req.Schedule.ToEntity(),
		IncludeDB:        req.IncludeDB,
		IncludeSpec:      req.IncludeSpec,
		TargetRepository: entity.ObjectID{ID: req.TargetRepository.ID},
		Notification:     req.Notification.ToEntity(),
	}
	if req.IncludeSpec {
		backup.SpecSecrets = string(req.SpecSecrets)
		if req.SpecSecrets == specmodel.SecretsModeEncrypted {
			backup.SpecPassphrase = entity.NewEncryptedField(req.SpecPassphrase)
		}
	}
	return backup
}

// KeepMaskedSecrets restores the stored values for the secrets the request only
// carries as the masked placeholder the GET response substitutes for them.
func (req *SystemBackupBaseReq) KeepMaskedSecrets(backup, current *entity.SystemBackup) {
	if current == nil {
		return
	}
	if basedto.IsMaskedSecret(req.SpecPassphrase) && backup.SpecSecrets == string(specmodel.SecretsModeEncrypted) {
		backup.SpecPassphrase = current.SpecPassphrase
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
	// A disabled backup is not asked for what it would take, nor where.
	active := req.Status == base.SettingStatusActive
	res = append(res, basedto.ValidateCond(!active || req.IncludeDB || req.IncludeSpec, field+"includeDB")...)
	if req.IncludeSpec {
		res = append(res, basedto.ValidateStrIn(&req.SpecSecrets, true, specmodel.AllSecretsModes,
			field+"specSecrets")...)
		res = append(res, basedto.ValidateStr(&req.SpecPassphrase,
			req.SpecSecrets == specmodel.SecretsModeEncrypted, 1, specPassphraseMaxLen, field+"specPassphrase")...)
	}
	res = append(res, basedto.ValidateObjectIDReq(&req.TargetRepository, active, field+"targetRepository")...)
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
func (req *UpdateSystemBackupReq) Validate() hperrors.ValidationErrors {
	validators := make([]vld.Validator, 0, 10) //nolint:mnd
	validators = append(validators, req.UpdateUniqueSettingReq.Validate()...)
	validators = append(validators, req.validate("")...)
	return hperrors.NewValidationErrors(vld.Validate(validators...))
}

type UpdateSystemBackupResp struct {
	Meta *basedto.Meta `json:"meta"`
}
