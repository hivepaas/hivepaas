package systembackupdto

import (
	"time"

	vld "github.com/tiendc/go-validator"

	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/basedto"
	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
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

func (req *GetSystemBackupReq) Validate() hperrors.ValidationErrors {
	validators := make([]vld.Validator, 0, 10) //nolint:mnd
	validators = append(validators, req.GetUniqueSettingReq.Validate()...)
	return hperrors.NewValidationErrors(vld.Validate(validators...))
}

type GetSystemBackupResp struct {
	Meta *basedto.Meta     `json:"meta"`
	Data *SystemBackupResp `json:"data"`
}

type SystemBackupResp struct {
	*settings.BaseSettingResp
	Schedule    *ScheduleResp `json:"schedule"`
	IncludeDB   bool          `json:"includeDB"`
	IncludeSpec bool          `json:"includeSpec"`
	SpecSecrets string        `json:"specSecrets,omitempty"`
	// SpecPassphrase comes masked once it is stored.
	SpecPassphrase   string                             `json:"specPassphrase,omitempty"`
	TargetRepository *settings.BaseSettingResp          `json:"targetRepository,omitempty"`
	Notification     *basedto.BaseEventNotificationResp `json:"notification"`
	SecretMasked     bool                               `json:"secretMasked,omitempty"`

	// Calculated fields
	NextRuns []time.Time `json:"nextRuns"`
}

type ScheduleResp struct {
	CronExpr    string            `json:"cronExpr,omitempty"` // cronExpr and interval are mutually exclusive
	Interval    timeutil.Duration `json:"interval,omitempty"`
	InitialTime time.Time         `json:"initialTime"`
}

func TransformSystemBackup(
	setting *entity.Setting,
	refObjects *entity.RefObjects,
) (resp *SystemBackupResp, err error) {
	config := setting.MustAsSystemBackup()
	resp = &SystemBackupResp{
		IncludeDB:   config.IncludeDB,
		IncludeSpec: config.IncludeSpec,
		SpecSecrets: config.SpecSecrets,
	}
	if err = copier.Copy(&resp.Schedule, &config.Schedule); err != nil {
		return nil, hperrors.Wrap(err)
	}

	resp.BaseSettingResp, err = settings.TransformSettingBase(setting)
	if err != nil {
		return nil, hperrors.Wrap(err)
	}

	resp.SpecPassphrase = config.SpecPassphrase.String()
	resp.SecretMasked = config.SpecPassphrase.IsEncrypted() || resp.Inherited
	if resp.SecretMasked && !config.SpecPassphrase.IsEmpty() {
		resp.SpecPassphrase = basedto.MaskedSecret
	}

	if refObjects == nil {
		refObjects = &entity.RefObjects{}
	}
	if config.TargetRepository.ID != "" {
		repoResp, _ := settings.TransformSettingBase(refObjects.RefSettings[config.TargetRepository.ID])
		if repoResp == nil {
			repoResp = settings.NewMissingSetting(config.TargetRepository.ID, base.SettingTypeBackupRepo)
		}
		resp.TargetRepository = repoResp
	}

	resp.Notification = basedto.TransformBaseEventNotification(config.Notification, refObjects)

	// Add next runs
	resp.NextRuns, _ = config.Schedule.CalcNextRuns(time.Now(), 5) //nolint

	return resp, nil
}
