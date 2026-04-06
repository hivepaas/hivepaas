package networkdto

import (
	vld "github.com/tiendc/go-validator"

	"github.com/localpaas/localpaas/localpaas_app/apperrors"
	"github.com/localpaas/localpaas/localpaas_app/basedto"
)

type GetNetworkInspectionReq struct {
	NetworkID string `json:"-"`
	ProjectID string `json:"-"`
}

func NewGetNetworkInspectionReq() *GetNetworkInspectionReq {
	return &GetNetworkInspectionReq{}
}

func (req *GetNetworkInspectionReq) Validate() apperrors.ValidationErrors {
	var validators []vld.Validator
	validators = append(validators, basedto.ValidateStr(&req.NetworkID, true, 1, networkIDMaxLen, "networkID")...)
	validators = append(validators, basedto.ValidateID(&req.ProjectID, false, "projectID")...)
	return apperrors.NewValidationErrors(vld.Validate(validators...))
}

type GetNetworkInspectionResp struct {
	Meta *basedto.Meta `json:"meta"`
	Data string        `json:"data"`
}
