package schedjobdto

import (
	"strings"
	"time"

	vld "github.com/tiendc/go-validator"
	"github.com/tiendc/gofn"

	"github.com/hivepaas/hivepaas/hivepaas_app/apperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/basedto"
	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/timeutil"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/settings"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/settings/commandtemplateuc/commandtemplatedto"
)

const (
	maxRetryCount = 100
	maxRetryDelay = timeutil.Duration(time.Hour * 24)
	maxTimeout    = timeutil.Duration(time.Hour * 24)

	encryptionMinLen = 6
	encryptionMaxLen = 50
)

type CreateSchedJobReq struct {
	settings.CreateSettingReq
	*SchedJobBaseReq
}

type SchedJobBaseReq struct {
	Name               string                                     `json:"name"`
	JobType            base.SchedJobType                          `json:"jobType"`
	Schedule           *ScheduleReq                               `json:"schedule"`
	App                basedto.ObjectIDReq                        `json:"app"`
	Priority           base.TaskPriority                          `json:"priority"`
	MaxRetry           int                                        `json:"maxRetry"`
	RetryDelay         timeutil.Duration                          `json:"retryDelay"`
	RetryDelayIncr     timeutil.Duration                          `json:"retryDelayIncr"`
	RetryBackoff       bool                                       `json:"retryBackoff"`
	RetryBackoffJitter timeutil.Duration                          `json:"retryBackoffJitter"`
	RetryDelayMax      timeutil.Duration                          `json:"retryDelayMax"`
	Timeout            timeutil.Duration                          `json:"timeout"`
	ControlDisabled    bool                                       `json:"controlDisabled"`
	Command            *commandtemplatedto.CommandTemplateBaseReq `json:"command"`
	CommandOutput      *CommandOutputReq                          `json:"commandOutput"`
	Notification       *basedto.BaseEventNotificationReq          `json:"notification"`
}

func (req *SchedJobBaseReq) ToEntity() *entity.SchedJob {
	res := &entity.SchedJob{
		JobType:            req.JobType,
		Schedule:           req.Schedule.ToEntity(),
		App:                entity.ObjectID{ID: req.App.ID},
		Priority:           req.Priority,
		MaxRetry:           req.MaxRetry,
		RetryDelay:         req.RetryDelay,
		RetryDelayIncr:     req.RetryDelayIncr,
		RetryBackoffJitter: req.RetryBackoffJitter,
		RetryDelayMax:      req.RetryDelayMax,
		Timeout:            req.Timeout,
		ControlDisabled:    req.ControlDisabled,
		Notification:       req.Notification.ToEntity(),
	}
	if req.JobType == base.SchedJobTypeContainerCommand {
		res.Command = req.Command.ToEntity()
		res.CommandOutput = req.CommandOutput.ToEntity()
	}
	return res
}

func (req *SchedJobBaseReq) modifyRequest() error {
	req.Name = strings.TrimSpace(req.Name)
	req.Priority = gofn.Coalesce(req.Priority, base.TaskPriorityDefault)
	if req.Schedule != nil {
		req.Schedule.CronExpr = strings.TrimSpace(req.Schedule.CronExpr)
		if req.Schedule.InitialTime.IsZero() {
			req.Schedule.InitialTime = timeutil.NowUTC()
		}
		req.Schedule.InitialTime = req.Schedule.InitialTime.Truncate(time.Second)
	}
	if req.RetryBackoff && req.RetryBackoffJitter <= 0 {
		req.RetryBackoffJitter = timeutil.Duration(time.Second)
	}
	if req.Command != nil {
		req.Command.Name = "-"
		req.Command.Kind = ""
		if err := req.Command.ModifyRequest(); err != nil {
			return apperrors.Wrap(err)
		}
	}
	if req.CommandOutput != nil && req.CommandOutput.PipeToApp != nil {
		req.CommandOutput.PipeToApp.Command.Name = "-"
		req.CommandOutput.PipeToApp.Command.Kind = ""
		if err := req.CommandOutput.PipeToApp.Command.ModifyRequest(); err != nil {
			return apperrors.Wrap(err)
		}
	}
	return nil
}

func (req *SchedJobBaseReq) validate(field string) (res []vld.Validator) {
	if field != "" {
		field += "."
	}
	res = append(res, basedto.ValidateStr(&req.Name, true, 1, base.SettingNameMaxLen, field+"name")...)
	res = append(res, basedto.ValidateStrIn(&req.JobType, true, base.AllSchedJobTypes, field+"jobType")...)
	res = append(res, req.Schedule.validate(field+"schedule")...)
	res = append(res, basedto.ValidateObjectIDReq(&req.App, false, field+"app")...)
	res = append(res, basedto.ValidateStrIn(&req.Priority, true, base.AllTaskPriorities, field+"priority")...)
	res = append(res, basedto.ValidateNumber(&req.MaxRetry, false, 1, maxRetryCount, field+"maxRetry")...)
	res = append(res, basedto.ValidateDuration(&req.RetryDelay, false, 1, maxRetryDelay, field+"retryDelay")...)
	res = append(res, basedto.ValidateDuration(&req.Timeout, false, 1, maxTimeout, field+"timeout")...)
	res = append(res, req.Command.Validate(field+"command")...)
	res = append(res, req.CommandOutput.validate(req.App.ID, field+"commandOutput")...)
	res = append(res, req.Notification.Validate(field+"notification")...)
	return res
}

type ScheduleReq struct {
	CronExpr    string            `json:"cronExpr"` // cronExpr and interval are mutually exclusive
	Interval    timeutil.Duration `json:"interval"`
	InitialTime time.Time         `json:"initialTime"`
	EndTime     time.Time         `json:"endTime"`
}

func (req *ScheduleReq) ToEntity() *entity.SchedJobSchedule {
	if req == nil {
		return nil
	}
	return &entity.SchedJobSchedule{
		CronExpr:    req.CronExpr,
		Interval:    req.Interval,
		InitialTime: req.InitialTime,
		EndTime:     req.EndTime,
	}
}

func (req *ScheduleReq) validate(field string) (res []vld.Validator) {
	if field != "" {
		field += "."
	}
	res = append(res, basedto.ValidateCond(req.ToEntity().IsValid() == nil, field+"cronExpr|interval")...)
	res = append(res, basedto.ValidateTime(&req.InitialTime, true,
		timeutil.NowUTC().Add(-timeutil.Dur365Days), time.Time{}, field+"initialTime")...)
	return res
}

type CommandOutputReq struct {
	Enabled    bool                        `json:"enabled"`
	SaveToFile *CommandOutputSaveToFileReq `json:"saveToFile"`
	PipeToApp  *CommandOutputPipeToAppReq  `json:"pipeToApp"`
}

func (req *CommandOutputReq) ToEntity() *entity.SchedJobCommandOutput {
	if req == nil {
		return nil
	}
	return &entity.SchedJobCommandOutput{
		Enabled:    req.Enabled,
		SaveToFile: req.SaveToFile.ToEntity(),
		PipeToApp:  req.PipeToApp.ToEntity(),
	}
}

func (req *CommandOutputReq) validate(appID string, field string) (res []vld.Validator) {
	if req == nil {
		return nil
	}
	if field != "" {
		field += "."
	}
	if req.Enabled {
		res = append(res, basedto.ValidateCond((req.SaveToFile != nil && req.PipeToApp == nil) ||
			(req.SaveToFile == nil && req.PipeToApp != nil), field+"saveToFile|pipeToApp")...)
	}
	if req.Enabled && req.PipeToApp != nil {
		res = append(res, basedto.ValidateCond(req.PipeToApp.TargetApp.ID != appID,
			field+"pipeToApp.targetApp.id")...)
	}
	res = append(res, req.SaveToFile.validate(field+"saveToFile")...)
	res = append(res, req.PipeToApp.validate(field+"pipeToApp")...)
	return res
}

type CommandOutputSaveToFileReq struct {
	FileName          string                      `json:"fileName"`
	FilePath          string                      `json:"filePath"`
	FileKind          base.FileKind               `json:"fileKind"`
	Storage           CommandOutputFileStorageReq `json:"storage"`
	CompressionFormat base.FileCompressionFormat  `json:"compressionFormat"`
	EncryptionFormat  base.FileEncryptionFormat   `json:"encryptionFormat"`
	EncryptionSecret  string                      `json:"encryptionSecret"`
}

func (req *CommandOutputSaveToFileReq) ToEntity() *entity.SchedJobCommandOutputSaveToFile {
	if req == nil {
		return nil
	}
	return &entity.SchedJobCommandOutputSaveToFile{
		FileName:          req.FileName,
		FilePath:          req.FilePath,
		FileKind:          req.FileKind,
		Storage:           *req.Storage.ToEntity(),
		CompressionFormat: req.CompressionFormat,
		EncryptionFormat:  req.EncryptionFormat,
		EncryptionSecret:  entity.NewEncryptedField(req.EncryptionSecret),
	}
}

func (req *CommandOutputSaveToFileReq) validate(field string) (res []vld.Validator) {
	if req == nil {
		return nil
	}
	if field != "" {
		field += "."
	}
	res = append(res, req.Storage.validate(field+"storage")...)
	res = append(res, basedto.ValidateStrIn(&req.CompressionFormat, false, base.AllFileCompressionFormats,
		field+"compressionFormat")...)
	res = append(res, basedto.ValidateStrIn(&req.EncryptionFormat, false, base.AllFileEncryptionFormats,
		field+"encryptionFormat")...)
	res = append(res, basedto.ValidateStr(&req.EncryptionSecret, req.EncryptionFormat != "",
		encryptionMinLen, encryptionMaxLen, field+"encryptionSecret")...)
	return res
}

type CommandOutputFileStorageReq struct {
	ID     string `json:"id"`
	Bucket string `json:"bucket"`
}

func (req *CommandOutputFileStorageReq) ToEntity() *entity.SchedJobCommandOutputFileStorage {
	if req == nil {
		return nil
	}
	return &entity.SchedJobCommandOutputFileStorage{
		ID:     req.ID,
		Bucket: req.Bucket,
	}
}

func (req *CommandOutputFileStorageReq) validate(field string) (res []vld.Validator) {
	if req == nil {
		return nil
	}
	if field != "" {
		field += "."
	}
	res = append(res, basedto.ValidateID(&req.ID, false, field+"id")...)
	return res
}

type CommandOutputPipeToAppReq struct {
	TargetApp basedto.ObjectIDReq                        `json:"targetApp"`
	Command   *commandtemplatedto.CommandTemplateBaseReq `json:"command"`
}

func (req *CommandOutputPipeToAppReq) ToEntity() *entity.SchedJobCommandOutputPipeToApp {
	if req == nil {
		return nil
	}
	return &entity.SchedJobCommandOutputPipeToApp{
		TargetApp: *req.TargetApp.ToEntity(),
		Command:   req.Command.ToEntity(),
	}
}

func (req *CommandOutputPipeToAppReq) validate(field string) (res []vld.Validator) {
	if req == nil {
		return nil
	}
	if field != "" {
		field += "."
	}
	res = append(res, basedto.ValidateObjectIDReq(&req.TargetApp, true, field+"targetApp")...)
	res = append(res, req.Command.Validate(field+"command")...)
	return res
}

func NewCreateSchedJobReq() *CreateSchedJobReq {
	return &CreateSchedJobReq{}
}

func (req *CreateSchedJobReq) ModifyRequest() error {
	return req.modifyRequest()
}

// Validate implements interface basedto.ReqValidator
func (req *CreateSchedJobReq) Validate() apperrors.ValidationErrors {
	var validators []vld.Validator
	validators = append(validators, req.validate("")...)
	return apperrors.NewValidationErrors(vld.Validate(validators...))
}

type CreateSchedJobResp struct {
	Meta *basedto.Meta         `json:"meta"`
	Data *basedto.ObjectIDResp `json:"data"`
}
