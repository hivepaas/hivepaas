package backuprepocleanupdto

import (
	"time"

	vld "github.com/tiendc/go-validator"

	"github.com/hivepaas/hivepaas/hivepaas_app/basedto"
	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/copier"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/timeutil"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/settings"
)

type GetBackupRepoCleanupReq struct {
	settings.GetUniqueSettingReq
}

func NewGetBackupRepoCleanupReq() *GetBackupRepoCleanupReq {
	return &GetBackupRepoCleanupReq{}
}

func (req *GetBackupRepoCleanupReq) Validate() hperrors.ValidationErrors {
	validators := make([]vld.Validator, 0, 10) //nolint:mnd
	validators = append(validators, req.GetUniqueSettingReq.Validate()...)
	return hperrors.NewValidationErrors(vld.Validate(validators...))
}

type GetBackupRepoCleanupResp struct {
	Meta *basedto.Meta          `json:"meta"`
	Data *BackupRepoCleanupResp `json:"data"`
}

type BackupRepoCleanupResp struct {
	*settings.BaseSettingResp
	Schedule     *ScheduleResp                      `json:"schedule"`
	Notification *basedto.BaseEventNotificationResp `json:"notification"`

	// Calculated fields
	NextRuns []time.Time `json:"nextRuns"`
}

type ScheduleResp struct {
	CronExpr    string            `json:"cronExpr,omitempty"` // cronExpr and interval are mutually exclusive
	Interval    timeutil.Duration `json:"interval,omitempty"`
	InitialTime time.Time         `json:"initialTime"`
}

func TransformBackupRepoCleanup(
	setting *entity.Setting,
	refObjects *entity.RefObjects,
) (resp *BackupRepoCleanupResp, err error) {
	config := setting.MustAsBackupRepoCleanup()
	if err = copier.Copy(&resp, config); err != nil {
		return nil, hperrors.Wrap(err)
	}

	resp.BaseSettingResp, err = settings.TransformSettingBase(setting)
	if err != nil {
		return nil, hperrors.Wrap(err)
	}

	resp.Notification = basedto.TransformBaseEventNotification(config.Notification, refObjects)

	// Add next runs
	resp.NextRuns, _ = config.Schedule.CalcNextRuns(time.Now(), 5) //nolint

	return resp, nil
}
