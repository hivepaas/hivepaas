package appdto

import (
	vld "github.com/tiendc/go-validator"

	"github.com/hivepaas/hivepaas/hivepaas_app/apperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/basedto"
)

type GetTerminalInfoReq struct {
	ProjectID string `json:"-"`
	AppID     string `json:"-"`
}

func NewGetTerminalInfoReq() *GetTerminalInfoReq {
	return &GetTerminalInfoReq{}
}

func (req *GetTerminalInfoReq) Validate() apperrors.ValidationErrors {
	var validators []vld.Validator
	validators = append(validators, basedto.ValidateID(&req.ProjectID, true, "projectId")...)
	validators = append(validators, basedto.ValidateID(&req.AppID, true, "appId")...)
	return apperrors.NewValidationErrors(vld.Validate(validators...))
}

type GetTerminalInfoResp struct {
	Meta *basedto.Meta         `json:"meta"`
	Data *TerminalInfoDataResp `json:"data"`
}

type TerminalInfoDataResp struct {
	Enabled         bool     `json:"enabled"`
	SupportedShells []string `json:"supportedShells"`
}
