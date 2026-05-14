package lpappsettingsdto

import (
	"github.com/moby/moby/api/types/swarm"

	"github.com/localpaas/localpaas/localpaas_app/apperrors"
	"github.com/localpaas/localpaas/localpaas_app/basedto"
	"github.com/localpaas/localpaas/localpaas_app/entity"
	"github.com/localpaas/localpaas/localpaas_app/pkg/copier"
	"github.com/localpaas/localpaas/localpaas_app/pkg/timeutil"
	"github.com/localpaas/localpaas/localpaas_app/usecase/settings"
)

type GetServiceSettingsReq struct {
}

func NewGetServiceSettingsReq() *GetServiceSettingsReq {
	return &GetServiceSettingsReq{}
}

func (req *GetServiceSettingsReq) Validate() apperrors.ValidationErrors {
	return nil
}

type GetServiceSettingsResp struct {
	Meta *basedto.Meta        `json:"meta"`
	Data *ServiceSettingsResp `json:"data"`
}

type ServiceSettingsResp struct {
	*settings.BaseSettingResp
	AppSettings         LocalPaaSAppSettingsResp         `json:"appSettings"`
	WorkerSettings      LocalPaaSWorkerSettingsResp      `json:"workerSettings"`
	TaskSettings        LocalPaaSTaskSettingsResp        `json:"taskSettings"`
	HealthcheckSettings LocalPaaSHealthcheckSettingsResp `json:"healthcheckSettings"`
}

type LocalPaaSAppSettingsResp struct {
	Replicas int `json:"replicas"`
}

type LocalPaaSWorkerSettingsResp struct {
	Replicas           int  `json:"replicas"`
	Concurrency        int  `json:"concurrency"`
	RunWorkerInMainApp bool `json:"runWorkerInMainApp"`
}

type LocalPaaSTaskSettingsResp struct {
	TaskCheckInterval  timeutil.Duration `json:"taskCheckInterval"`
	TaskCreateInterval timeutil.Duration `json:"taskCreateInterval"`
}

type LocalPaaSHealthcheckSettingsResp struct {
	BaseInterval timeutil.Duration `json:"baseInterval"`
}

type ServiceSettingsTransformInput struct {
	Setting       *entity.Setting
	MainService   *swarm.Service
	WorkerService *swarm.Service
}

func TransformServiceSettings(
	input *ServiceSettingsTransformInput,
) (resp *ServiceSettingsResp, err error) {
	config := input.Setting.MustAsLocalPaaSService()
	if err = copier.Copy(&resp, config); err != nil {
		return nil, apperrors.Wrap(err)
	}
	resp.BaseSettingResp, err = settings.TransformSettingBase(input.Setting)
	if err != nil {
		return nil, apperrors.Wrap(err)
	}

	// Some dynamic info retrieved from the infra
	resp.AppSettings.Replicas = int(*input.MainService.Spec.Mode.Replicated.Replicas)      //nolint
	resp.WorkerSettings.Replicas = int(*input.WorkerService.Spec.Mode.Replicated.Replicas) //nolint

	return resp, nil
}
