package traefiksettingsdto

import (
	"github.com/moby/moby/api/types/swarm"

	"github.com/hivepaas/hivepaas/hivepaas_app/apperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/basedto"
	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/copier"
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
	AppSettings TraefikAppSettingsResp `json:"appSettings"`
}

type TraefikAppSettingsResp struct {
	Replicas int `json:"replicas"`
}

type ServiceSettingsTransformInput struct {
	Setting        *entity.Setting
	TraefikService *swarm.Service
}

func TransformServiceSettings(
	input *ServiceSettingsTransformInput,
) (resp *ServiceSettingsResp, err error) {
	config := input.Setting.MustAsTraefikService()
	if err = copier.Copy(&resp, config); err != nil {
		return nil, apperrors.New(err)
	}
	resp.BaseSettingResp, err = settings.TransformSettingBase(input.Setting)
	if err != nil {
		return nil, apperrors.New(err)
	}

	// Some dynamic info retrieved from the infra
	resp.AppSettings.Replicas = int(*input.TraefikService.Spec.Mode.Replicated.Replicas) //nolint

	return resp, nil
}
