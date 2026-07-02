package hpappsettingsdto

import (
	"github.com/moby/moby/api/types/swarm"

	"github.com/hivepaas/hivepaas/hivepaas_app/apperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/basedto"
	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/copier"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/timeutil"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/settings"
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
	AppSettings         HivePaaSAppSettingsResp         `json:"appSettings"`
	WorkerSettings      HivePaaSWorkerSettingsResp      `json:"workerSettings"`
	TaskSettings        HivePaaSTaskSettingsResp        `json:"taskSettings"`
	HealthcheckSettings HivePaaSHealthcheckSettingsResp `json:"healthcheckSettings"`
}

type HivePaaSAppSettingsResp struct {
	Replicas int `json:"replicas"`
}

type HivePaaSWorkerSettingsResp struct {
	Replicas           int  `json:"replicas"`
	Concurrency        int  `json:"concurrency"`
	RunWorkerInMainApp bool `json:"runWorkerInMainApp"`
}

type HivePaaSTaskSettingsResp struct {
	TaskCheckInterval  timeutil.Duration `json:"taskCheckInterval"`
	TaskCreateInterval timeutil.Duration `json:"taskCreateInterval"`
}

type HivePaaSHealthcheckSettingsResp struct {
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
	config := input.Setting.MustAsHivePaaSService()
	if err = copier.Copy(&resp, config); err != nil {
		return nil, apperrors.New(err)
	}
	resp.BaseSettingResp, err = settings.TransformSettingBase(input.Setting)
	if err != nil {
		return nil, apperrors.New(err)
	}

	// Some dynamic info retrieved from the infra
	resp.AppSettings.Replicas = int(*input.MainService.Spec.Mode.Replicated.Replicas)      //nolint
	resp.WorkerSettings.Replicas = int(*input.WorkerService.Spec.Mode.Replicated.Replicas) //nolint

	return resp, nil
}
