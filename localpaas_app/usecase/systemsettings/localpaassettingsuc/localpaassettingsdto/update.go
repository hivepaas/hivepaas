package localpaassettingsdto

import (
	vld "github.com/tiendc/go-validator"

	"github.com/localpaas/localpaas/localpaas_app/apperrors"
	"github.com/localpaas/localpaas/localpaas_app/basedto"
	"github.com/localpaas/localpaas/localpaas_app/entity"
	"github.com/localpaas/localpaas/localpaas_app/pkg/timeutil"
	"github.com/localpaas/localpaas/localpaas_app/usecase/settings"
)

type UpdateLocalPaaSSettingsReq struct {
	settings.UpdateUniqueSettingReq
	*LocalPaaSSettingsBaseReq
}

type LocalPaaSSettingsBaseReq struct {
	WorkerSettings      *LocalPaaSWorkerSettingsReq      `json:"workerSettings"`
	TaskSettings        *LocalPaaSTaskSettingsReq        `json:"taskSettings"`
	HealthcheckSettings *LocalPaaSHealthcheckSettingsReq `json:"healthcheckSettings"`
}

func (req *LocalPaaSSettingsBaseReq) ToEntity() *entity.LocalPaaSSettings {
	return &entity.LocalPaaSSettings{
		WorkerSettings:      req.WorkerSettings.ToEntity(),
		TaskSettings:        req.TaskSettings.ToEntity(),
		HealthcheckSettings: req.HealthcheckSettings.ToEntity(),
	}
}

// nolint
func (req *LocalPaaSSettingsBaseReq) validate(field string) (res []vld.Validator) {
	if field != "" {
		field += "."
	}
	// TODO: add validation
	return res
}

type LocalPaaSWorkerSettingsReq struct {
	Replicas           int  `json:"replicas,omitempty"`
	Concurrency        int  `json:"concurrency,omitempty"`
	RunWorkerInMainApp bool `json:"runWorkerInMainApp,omitempty"`
}

func (req *LocalPaaSWorkerSettingsReq) ToEntity() *entity.LocalPaaSWorkerSettings {
	return &entity.LocalPaaSWorkerSettings{
		Replicas:           req.Replicas,
		Concurrency:        req.Concurrency,
		RunWorkerInMainApp: req.RunWorkerInMainApp,
	}
}

type LocalPaaSTaskSettingsReq struct {
	TaskCheckInterval  timeutil.Duration `json:"taskCheckInterval"`
	TaskCreateInterval timeutil.Duration `json:"taskCreateInterval"`
}

func (req *LocalPaaSTaskSettingsReq) ToEntity() *entity.LocalPaaSTaskSettings {
	return &entity.LocalPaaSTaskSettings{
		TaskCheckInterval:  req.TaskCheckInterval,
		TaskCreateInterval: req.TaskCreateInterval,
	}
}

type LocalPaaSHealthcheckSettingsReq struct {
	BaseInterval timeutil.Duration `json:"baseInterval"`
}

func (req *LocalPaaSHealthcheckSettingsReq) ToEntity() *entity.LocalPaaSHealthcheckSettings {
	return &entity.LocalPaaSHealthcheckSettings{
		BaseInterval: req.BaseInterval,
	}
}

func NewUpdateLocalPaaSSettingsReq() *UpdateLocalPaaSSettingsReq {
	return &UpdateLocalPaaSSettingsReq{}
}

// Validate implements interface basedto.ReqValidator
func (req *UpdateLocalPaaSSettingsReq) Validate() apperrors.ValidationErrors {
	var validators []vld.Validator
	validators = append(validators, req.validate("")...)
	return apperrors.NewValidationErrors(vld.Validate(validators...))
}

type UpdateLocalPaaSSettingsResp struct {
	Meta *basedto.Meta `json:"meta"`
}
