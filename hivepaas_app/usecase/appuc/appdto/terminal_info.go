package appdto

import (
	vld "github.com/tiendc/go-validator"

	"github.com/hivepaas/hivepaas/hivepaas_app/basedto"
	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
)

type GetTerminalInfoReq struct {
	ProjectID    string `json:"-"`
	ProjectEnvID string `json:"-"`
	AppID        string `json:"-"`
}

func NewGetTerminalInfoReq() *GetTerminalInfoReq {
	return &GetTerminalInfoReq{}
}

func (req *GetTerminalInfoReq) Validate() hperrors.ValidationErrors {
	validators := make([]vld.Validator, 0, 10) //nolint:mnd
	validators = append(validators, basedto.ValidateID(&req.ProjectID, true, "projectId")...)
	validators = append(validators, basedto.ValidateID(&req.ProjectEnvID, true, "projectEnv")...)
	validators = append(validators, basedto.ValidateID(&req.AppID, true, "appId")...)
	return hperrors.NewValidationErrors(vld.Validate(validators...))
}

type GetTerminalInfoResp struct {
	Meta *basedto.Meta         `json:"meta"`
	Data *TerminalInfoDataResp `json:"data"`
}

type TerminalInfoDataResp struct {
	Enabled         bool     `json:"enabled"`
	SupportedShells []string `json:"supportedShells"`
}
