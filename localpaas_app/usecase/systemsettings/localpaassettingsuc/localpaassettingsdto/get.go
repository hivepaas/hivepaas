package localpaassettingsdto

import (
	vld "github.com/tiendc/go-validator"

	"github.com/localpaas/localpaas/localpaas_app/apperrors"
	"github.com/localpaas/localpaas/localpaas_app/basedto"
	"github.com/localpaas/localpaas/localpaas_app/entity"
	"github.com/localpaas/localpaas/localpaas_app/pkg/copier"
	"github.com/localpaas/localpaas/localpaas_app/pkg/timeutil"
	"github.com/localpaas/localpaas/localpaas_app/usecase/settings"
)

type GetLocalPaaSSettingsReq struct {
	settings.GetUniqueSettingReq
}

func NewGetLocalPaaSSettingsReq() *GetLocalPaaSSettingsReq {
	return &GetLocalPaaSSettingsReq{}
}

func (req *GetLocalPaaSSettingsReq) Validate() apperrors.ValidationErrors {
	var validators []vld.Validator
	validators = append(validators, req.GetUniqueSettingReq.Validate()...)
	return apperrors.NewValidationErrors(vld.Validate(validators...))
}

type GetLocalPaaSSettingsResp struct {
	Meta *basedto.Meta          `json:"meta"`
	Data *LocalPaaSSettingsResp `json:"data"`
}

type LocalPaaSSettingsResp struct {
	*settings.BaseSettingResp
	WorkerSettings      *LocalPaaSWorkerSettingsResp      `json:"workerSettings"`
	TaskSettings        *LocalPaaSTaskSettingsResp        `json:"taskSettings"`
	HealthcheckSettings *LocalPaaSHealthcheckSettingsResp `json:"healthcheckSettings"`
}

type LocalPaaSWorkerSettingsResp struct {
	Replicas           int  `json:"replicas,omitempty"`
	Concurrency        int  `json:"concurrency,omitempty"`
	RunWorkerInMainApp bool `json:"runWorkerInMainApp,omitempty"`
}

type LocalPaaSTaskSettingsResp struct {
	TaskCheckInterval  timeutil.Duration `json:"taskCheckInterval"`
	TaskCreateInterval timeutil.Duration `json:"taskCreateInterval"`
}

type LocalPaaSHealthcheckSettingsResp struct {
	BaseInterval timeutil.Duration `json:"baseInterval"`
}

func TransformLocalPaaSSettings(
	setting *entity.Setting,
	_ *entity.RefObjects,
) (resp *LocalPaaSSettingsResp, err error) {
	config := setting.MustAsLocalPaaSSettings()
	if err = copier.Copy(&resp, config); err != nil {
		return nil, apperrors.Wrap(err)
	}

	resp.BaseSettingResp, err = settings.TransformSettingBase(setting)
	if err != nil {
		return nil, apperrors.Wrap(err)
	}

	return resp, nil
}
