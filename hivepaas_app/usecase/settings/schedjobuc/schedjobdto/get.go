package schedjobdto

import (
	"time"

	vld "github.com/tiendc/go-validator"
	"github.com/tiendc/gofn"

	"github.com/hivepaas/hivepaas/hivepaas_app/apperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/basedto"
	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/copier"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/timeutil"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/appuc/appdto"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/settings"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/settings/commandtemplateuc/commandtemplatedto"
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
	JobType            base.SchedJobType                       `json:"jobType"`
	Schedule           *ScheduleResp                           `json:"schedule"`
	App                *basedto.NamedObjectResp                `json:"app"`
	Priority           base.TaskPriority                       `json:"priority"`
	MaxRetry           int                                     `json:"maxRetry"`
	RetryDelay         timeutil.Duration                       `json:"retryDelay"`
	RetryDelayIncr     timeutil.Duration                       `json:"retryDelayIncr,omitempty"`
	RetryBackoff       bool                                    `json:"retryBackoff,omitempty"`
	RetryBackoffJitter timeutil.Duration                       `json:"retryBackoffJitter,omitempty"`
	RetryDelayMax      timeutil.Duration                       `json:"retryDelayMax,omitempty"`
	Timeout            timeutil.Duration                       `json:"timeout"`
	ControlDisabled    bool                                    `json:"controlDisabled"`
	Command            *commandtemplatedto.CommandTemplateResp `json:"command"`
	CommandOutput      *CommandOutputResp                      `json:"commandOutput,omitempty"`
	Notification       *basedto.BaseEventNotificationResp      `json:"notification"`

	// Calculated fields
	NextRuns []time.Time `json:"nextRuns,omitempty"`
}

type ScheduleResp struct {
	CronExpr    string            `json:"cronExpr,omitempty"` // cronExpr and interval are mutually exclusive
	Interval    timeutil.Duration `json:"interval,omitempty"`
	InitialTime time.Time         `json:"initialTime"`
	EndTime     time.Time         `json:"endTime,omitzero"`
}

type CommandOutputResp struct {
	Enabled    bool                         `json:"enabled"`
	SaveToFile *CommandOutputSaveToFileResp `json:"saveToFile,omitempty"`
	PipeToApp  *CommandOutputPipeToAppResp  `json:"pipeToApp,omitempty"`
}

type CommandOutputSaveToFileResp struct {
	FileName          string                     `json:"fileName"`
	FilePath          string                     `json:"filePath"`
	FileKind          base.FileKind              `json:"fileKind"`
	Storage           *settings.BaseSettingResp  `json:"storage,omitempty"`
	CompressionFormat base.FileCompressionFormat `json:"compressionFormat,omitempty"`
	EncryptionFormat  base.FileEncryptionFormat  `json:"encryptionFormat,omitempty"`
	EncryptionSecret  string                     `json:"encryptionSecret,omitempty"`
}

func (resp *CommandOutputSaveToFileResp) CopyEncryptionSecret(field entity.EncryptedField) error {
	resp.EncryptionSecret = field.String()
	return nil
}

type CommandOutputPipeToAppResp struct {
	TargetApp *appdto.AppResp                         `json:"targetApp"`
	Command   *commandtemplatedto.CommandTemplateResp `json:"command"`
}

func TransformSchedJob(
	setting *entity.Setting,
	refObjects *entity.RefObjects,
	isListAPI bool,
) (resp *SchedJobResp, err error) {
	job := setting.MustAsSchedJob()
	if err = copier.Copy(&resp, job); err != nil {
		return nil, apperrors.New(err)
	}

	resp.BaseSettingResp, err = settings.TransformSettingBase(setting)
	if err != nil {
		return nil, apperrors.New(err)
	}

	if job.App.ID != "" {
		refApp := refObjects.RefApps[job.App.ID]
		if err = copier.Copy(&resp.App, refApp); err != nil {
			return nil, apperrors.New(err)
		}
	} else {
		resp.App = nil
	}

	if job.CommandOutput != nil { //nolint:nestif
		cmdOutput := job.CommandOutput
		cmdOutputResp := resp.CommandOutput

		if cmdOutput.SaveToFile != nil && cmdOutput.SaveToFile.Storage.ID != "" {
			targetStorage := refObjects.RefSettings[cmdOutput.SaveToFile.Storage.ID]
			cmdOutputResp.SaveToFile.Storage, err = settings.TransformSettingBase(targetStorage)
			if err != nil {
				return nil, apperrors.New(err)
			}
		} else {
			cmdOutputResp.SaveToFile = nil
		}

		if cmdOutput.PipeToApp != nil && cmdOutput.PipeToApp.TargetApp.ID != "" {
			targetApp := refObjects.RefApps[cmdOutput.PipeToApp.TargetApp.ID]
			cmdOutputResp.PipeToApp.TargetApp, err = appdto.TransformApp(targetApp, nil)
			if err != nil {
				return nil, apperrors.New(err)
			}
		} else {
			cmdOutputResp.PipeToApp = nil
		}
	}

	resp.Notification = basedto.TransformBaseEventNotification(job.Notification, refObjects)

	// Custom fields
	if !resp.RetryBackoff && resp.RetryBackoffJitter > 0 {
		resp.RetryBackoff = true
	}

	if setting.IsActive() {
		count := gofn.If(isListAPI, 1, 5) //nolint:mnd
		resp.NextRuns, _ = job.Schedule.CalcNextRuns(timeutil.NowUTC(), count)
	}

	return resp, nil
}
