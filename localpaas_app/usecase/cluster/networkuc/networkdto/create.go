package networkdto

import (
	vld "github.com/tiendc/go-validator"

	"github.com/localpaas/localpaas/localpaas_app/apperrors"
	"github.com/localpaas/localpaas/localpaas_app/basedto"
	"github.com/localpaas/localpaas/localpaas_app/entity"
	"github.com/localpaas/localpaas/localpaas_app/usecase/settings"
)

const (
	networkNameMaxLen = 100
)

type CreateNetworkReq struct {
	settings.CreateSettingReq
	*CreateNetworkBaseReq
}

type CreateNetworkBaseReq struct {
	Name       string            `json:"name"`
	Driver     string            `json:"driver"`
	EnableIPv4 bool              `json:"enableIPv4"`
	EnableIPv6 bool              `json:"enableIPv6"`
	Internal   bool              `json:"internal"`
	Attachable bool              `json:"attachable"`
	Options    map[string]string `json:"options"`
	Labels     map[string]string `json:"labels"`
}

func (req *CreateNetworkBaseReq) validate(field string) (res []vld.Validator) {
	if req == nil {
		return res
	}
	if field != "" {
		field += "."
	}
	res = append(res, basedto.ValidateStr(&req.Name, true, 1, networkNameMaxLen, field+"name")...)
	return res
}

func (req *CreateNetworkBaseReq) ToEntity() *entity.ClusterNetwork {
	return &entity.ClusterNetwork{
		Name: req.Name,
	}
}

func NewCreateNetworkReq() *CreateNetworkReq {
	return &CreateNetworkReq{}
}

// Validate implements interface basedto.ReqValidator
func (req *CreateNetworkReq) Validate() apperrors.ValidationErrors {
	var validators []vld.Validator
	validators = append(validators, req.validate("")...)
	return apperrors.NewValidationErrors(vld.Validate(validators...))
}

type CreateNetworkResp struct {
	Meta *basedto.Meta         `json:"meta"`
	Data *basedto.ObjectIDResp `json:"data"`
}
