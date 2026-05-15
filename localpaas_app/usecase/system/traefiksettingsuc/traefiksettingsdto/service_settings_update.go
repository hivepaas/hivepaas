package traefiksettingsdto

import (
	vld "github.com/tiendc/go-validator"

	"github.com/localpaas/localpaas/localpaas_app/apperrors"
	"github.com/localpaas/localpaas/localpaas_app/basedto"
	"github.com/localpaas/localpaas/localpaas_app/entity"
)

const (
	traefikServiceReplicasMin = 1
	traefikServiceReplicasMax = 100
)

type UpdateServiceSettingsReq struct {
	*ServiceSettingsBaseReq
	UpdateVer int `json:"updateVer"`
}

type ServiceSettingsBaseReq struct {
	AppSettings TraefikAppSettingsReq `json:"appSettings"`
}

func (req *ServiceSettingsBaseReq) ToEntity() *entity.TraefikService {
	return &entity.TraefikService{
		AppSettings: *req.AppSettings.ToEntity(),
	}
}

func (req *ServiceSettingsBaseReq) validate(field string) (res []vld.Validator) {
	if field != "" {
		field += "."
	}
	res = append(res, req.AppSettings.validate(field+"appSettings")...)
	return res
}

type TraefikAppSettingsReq struct {
	Replicas int `json:"replicas"`
}

func (req *TraefikAppSettingsReq) ToEntity() *entity.TraefikAppSettings {
	return &entity.TraefikAppSettings{
		Replicas: req.Replicas,
	}
}

func (req *TraefikAppSettingsReq) validate(field string) (res []vld.Validator) {
	if field != "" {
		field += "."
	}
	res = append(res, basedto.ValidateNumber(&req.Replicas, true, traefikServiceReplicasMin,
		traefikServiceReplicasMax, field+"replicas")...)
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
