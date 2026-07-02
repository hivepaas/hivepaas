package hpappsettingsdto

import (
	"time"

	vld "github.com/tiendc/go-validator"

	"github.com/hivepaas/hivepaas/hivepaas_app/apperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/basedto"
	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/timeutil"
)

const (
	mainServiceReplicasMin = 1
	mainServiceReplicasMax = 100

	workerServiceReplicasMin    = 0
	workerServiceReplicasMax    = 100
	workerServiceConcurrencyMin = 1
	workerServiceConcurrencyMax = 100

	taskCheckIntervalMin  = timeutil.Duration(30 * time.Second)
	taskCheckIntervalMax  = timeutil.Duration(24 * time.Hour)
	taskCreateIntervalMin = timeutil.Duration(30 * time.Second)
	taskCreateIntervalMax = timeutil.Duration(24 * time.Hour)

	healthcheckBaseIntervalMin = timeutil.Duration(5 * time.Second)
	healthcheckBaseIntervalMax = timeutil.Duration(24 * time.Hour)
)

type UpdateServiceSettingsReq struct {
	*ServiceSettingsBaseReq
	UpdateVer int `json:"updateVer"`
}

type ServiceSettingsBaseReq struct {
	AppSettings         HivePaaSAppSettingsReq         `json:"appSettings"`
	WorkerSettings      HivePaaSWorkerSettingsReq      `json:"workerSettings"`
	TaskSettings        HivePaaSTaskSettingsReq        `json:"taskSettings"`
	HealthcheckSettings HivePaaSHealthcheckSettingsReq `json:"healthcheckSettings"`
}

func (req *ServiceSettingsBaseReq) ToEntity() *entity.HivePaaSService {
	return &entity.HivePaaSService{
		AppSettings:         *req.AppSettings.ToEntity(),
		WorkerSettings:      *req.WorkerSettings.ToEntity(),
		TaskSettings:        *req.TaskSettings.ToEntity(),
		HealthcheckSettings: *req.HealthcheckSettings.ToEntity(),
	}
}

func (req *ServiceSettingsBaseReq) validate(field string) (res []vld.Validator) {
	if field != "" {
		field += "."
	}
	res = append(res, req.AppSettings.validate(field+"appSettings")...)
	res = append(res, req.WorkerSettings.validate(field+"workerSettings")...)
	res = append(res, req.TaskSettings.validate(field+"taskSettings")...)
	res = append(res, req.HealthcheckSettings.validate(field+"healthcheckSettings")...)
	return res
}

type HivePaaSAppSettingsReq struct {
	Replicas int `json:"replicas"`
}

func (req *HivePaaSAppSettingsReq) ToEntity() *entity.HivePaaSAppSettings {
	return &entity.HivePaaSAppSettings{
		Replicas: req.Replicas,
	}
}

func (req *HivePaaSAppSettingsReq) validate(field string) (res []vld.Validator) {
	if field != "" {
		field += "."
	}
	res = append(res, basedto.ValidateNumber(&req.Replicas, true, mainServiceReplicasMin,
		mainServiceReplicasMax, field+"replicas")...)
	return res
}

type HivePaaSWorkerSettingsReq struct {
	Replicas           int  `json:"replicas"`
	Concurrency        int  `json:"concurrency"`
	RunWorkerInMainApp bool `json:"runWorkerInMainApp"`
}

func (req *HivePaaSWorkerSettingsReq) ToEntity() *entity.HivePaaSWorkerSettings {
	return &entity.HivePaaSWorkerSettings{
		Replicas:           req.Replicas,
		Concurrency:        req.Concurrency,
		RunWorkerInMainApp: req.RunWorkerInMainApp,
	}
}

func (req *HivePaaSWorkerSettingsReq) validate(field string) (res []vld.Validator) {
	if field != "" {
		field += "."
	}
	res = append(res, basedto.ValidateNumber(&req.Replicas, true, workerServiceReplicasMin,
		workerServiceReplicasMax, field+"replicas")...)
	res = append(res, basedto.ValidateNumber(&req.Concurrency, true, workerServiceConcurrencyMin,
		workerServiceConcurrencyMax, field+"concurrency")...)
	if req.Replicas == 0 && !req.RunWorkerInMainApp {
		res = append(res, vld.Must(false).OnError(
			vld.SetField(field+"runWorkerInMainApp", nil),
			vld.SetCustomKey("ERR_VLD_VALUE_INVALID"),
		))
	}
	return res
}

type HivePaaSTaskSettingsReq struct {
	TaskCheckInterval  timeutil.Duration `json:"taskCheckInterval"`
	TaskCreateInterval timeutil.Duration `json:"taskCreateInterval"`
}

func (req *HivePaaSTaskSettingsReq) ToEntity() *entity.HivePaaSTaskSettings {
	return &entity.HivePaaSTaskSettings{
		TaskCheckInterval:  req.TaskCheckInterval,
		TaskCreateInterval: req.TaskCreateInterval,
	}
}

func (req *HivePaaSTaskSettingsReq) validate(field string) (res []vld.Validator) {
	if field != "" {
		field += "."
	}
	res = append(res, basedto.ValidateDuration(&req.TaskCheckInterval, true, taskCheckIntervalMin,
		taskCheckIntervalMax, field+"taskCheckInterval")...)
	res = append(res, basedto.ValidateDuration(&req.TaskCreateInterval, true, taskCreateIntervalMin,
		taskCreateIntervalMax, field+"taskCreateInterval")...)
	return res
}

type HivePaaSHealthcheckSettingsReq struct {
	BaseInterval timeutil.Duration `json:"baseInterval"`
}

func (req *HivePaaSHealthcheckSettingsReq) ToEntity() *entity.HivePaaSHealthcheckSettings {
	return &entity.HivePaaSHealthcheckSettings{
		BaseInterval: req.BaseInterval,
	}
}

func (req *HivePaaSHealthcheckSettingsReq) validate(field string) (res []vld.Validator) {
	if field != "" {
		field += "."
	}
	res = append(res, basedto.ValidateDuration(&req.BaseInterval, true, healthcheckBaseIntervalMin,
		healthcheckBaseIntervalMax, field+"baseInterval")...)
	return res
}

func NewUpdateServiceSettingsReq() *UpdateServiceSettingsReq {
	return &UpdateServiceSettingsReq{}
}

// Validate implements interface basedto.ReqValidator
func (req *UpdateServiceSettingsReq) Validate() apperrors.ValidationErrors {
	var validators []vld.Validator
	validators = append(validators, req.validate("")...)
	return apperrors.NewValidationErrors(vld.Validate(validators...))
}

type UpdateServiceSettingsResp struct {
	Meta *basedto.Meta `json:"meta"`
}
