package schedjobdto

import (
	"time"

	vld "github.com/tiendc/go-validator"
	"github.com/tiendc/gofn"

	"github.com/localpaas/localpaas/localpaas_app/apperrors"
	"github.com/localpaas/localpaas/localpaas_app/base"
	"github.com/localpaas/localpaas/localpaas_app/basedto"
	"github.com/localpaas/localpaas/localpaas_app/entity"
	"github.com/localpaas/localpaas/localpaas_app/pkg/copier"
	"github.com/localpaas/localpaas/localpaas_app/pkg/timeutil"
	"github.com/localpaas/localpaas/localpaas_app/usecase/settings"
)

type GetSchedJobReq struct {
	settings.GetSettingReq
}

func NewGetSchedJobReq() *GetSchedJobReq {
	return &GetSchedJobReq{}
}

func (req *GetSchedJobReq) Validate() apperrors.ValidationErrors {
	var validators []vld.Validator
	validators = append(validators, req.GetSettingReq.Validate()...)
	return apperrors.NewValidationErrors(vld.Validate(validators...))
}

type GetSchedJobResp struct {
	Meta *basedto.Meta `json:"meta"`
	Data *SchedJobResp `json:"data"`
}

type SchedJobResp struct {
	*settings.BaseSettingResp
	JobType         base.SchedJobType                  `json:"jobType"`
	Schedule        *ScheduleResp                      `json:"schedule"`
	App             *basedto.NamedObjectResp           `json:"app"`
	Priority        base.TaskPriority                  `json:"priority"`
	MaxRetry        int                                `json:"maxRetry"`
	RetryDelay      timeutil.Duration                  `json:"retryDelay"`
	Timeout         timeutil.Duration                  `json:"timeout"`
	ControlDisabled bool                               `json:"controlDisabled"`
	Command         *ContainerCommandResp              `json:"command"`
	Notification    *basedto.BaseEventNotificationResp `json:"notification"`

	// Calculated fields
	NextRuns []time.Time `json:"nextRuns,omitempty"`
}

type ScheduleResp struct {
	CronExpr    string            `json:"cronExpr,omitempty"` // cronExpr and interval are mutually exclusive
	Interval    timeutil.Duration `json:"interval,omitempty"`
	InitialTime time.Time         `json:"initialTime"`
}

type ContainerCommandResp struct {
	RunInShell string                 `json:"runInShell"`
	Command    string                 `json:"command"`
	WorkingDir string                 `json:"workingDir"`
	EnvVars    []*basedto.EnvVarResp  `json:"envVars"`
	ArgGroups  []*CommandArgGroupResp `json:"argGroups"`
}

type CommandArgGroupResp struct {
	Enabled   bool              `json:"enabled"`
	ExportEnv string            `json:"exportEnv"`
	Separator string            `json:"separator"`
	Args      []*CommandArgResp `json:"args"`
}

type CommandArgResp struct {
	Use   bool   `json:"use"`
	Name  string `json:"name"`
	Value string `json:"value"`
}

func TransformSchedJob(
	setting *entity.Setting,
	refObjects *entity.RefObjects,
	isListAPI bool,
) (resp *SchedJobResp, err error) {
	job := setting.MustAsSchedJob()
	if err = copier.Copy(&resp, job); err != nil {
		return nil, apperrors.Wrap(err)
	}

	resp.BaseSettingResp, err = settings.TransformSettingBase(setting)
	if err != nil {
		return nil, apperrors.Wrap(err)
	}

	if job.App.ID != "" {
		refApp := refObjects.RefApps[job.App.ID]
		if err = copier.Copy(&resp.App, refApp); err != nil {
			return nil, apperrors.Wrap(err)
		}
	}

	resp.Notification = basedto.TransformBaseEventNotification(job.Notification, refObjects)

	if setting.IsActive() {
		count := gofn.If(isListAPI, 1, 5) //nolint:mnd
		resp.NextRuns, _ = job.Schedule.CalcNextRuns(timeutil.NowUTC(), count)
	}

	return resp, nil
}
