package cronjobdto

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

type GetCronJobReq struct {
	settings.GetSettingReq
}

func NewGetCronJobReq() *GetCronJobReq {
	return &GetCronJobReq{}
}

func (req *GetCronJobReq) Validate() apperrors.ValidationErrors {
	var validators []vld.Validator
	validators = append(validators, req.GetSettingReq.Validate()...)
	return apperrors.NewValidationErrors(vld.Validate(validators...))
}

type GetCronJobResp struct {
	Meta *basedto.Meta `json:"meta"`
	Data *CronJobResp  `json:"data"`
}

type CronJobResp struct {
	*settings.BaseSettingResp
	CronType        base.CronJobType                   `json:"cronType"`
	Schedule        *ScheduleResp                      `json:"schedule"`
	App             *basedto.NamedObjectResp           `json:"app"`
	Priority        base.TaskPriority                  `json:"priority"`
	MaxRetry        int                                `json:"maxRetry"`
	RetryDelay      timeutil.Duration                  `json:"retryDelay"`
	Timeout         timeutil.Duration                  `json:"timeout"`
	ControlDisabled bool                               `json:"controlDisabled"`
	Command         *ContainerCommandResp              `json:"command"`
	Notification    *basedto.BaseEventNotificationResp `json:"notification"`
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
	ExportEnv string            `json:"exportEnv"`
	Separator string            `json:"separator"`
	Args      []*CommandArgResp `json:"args"`
}

type CommandArgResp struct {
	Use   bool   `json:"use"`
	Name  string `json:"name"`
	Value string `json:"value"`
}

func TransformCronJob(
	setting *entity.Setting,
	refObjects *entity.RefObjects,
) (resp *CronJobResp, err error) {
	config := setting.MustAsCronJob()
	if err = copier.Copy(&resp, config); err != nil {
		return nil, apperrors.Wrap(err)
	}

	resp.BaseSettingResp, err = settings.TransformSettingBase(setting)
	if err != nil {
		return nil, apperrors.Wrap(err)
	}

	if config.App.ID != "" {
		refApp := refObjects.RefApps[config.App.ID]
		if err = copier.Copy(&resp.App, refApp); err != nil {
			return nil, apperrors.Wrap(err)
		}
	}

	resp.Notification = basedto.TransformBaseEventNotification(config.Notification, refObjects)
	return resp, nil
}
