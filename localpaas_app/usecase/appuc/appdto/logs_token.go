package appdto

import (
	vld "github.com/tiendc/go-validator"

	"github.com/localpaas/localpaas/localpaas_app/apperrors"
	"github.com/localpaas/localpaas/localpaas_app/basedto"
)

type GetAppLogsTokenReq struct {
	ProjectID string `json:"-"`
	AppID     string `json:"-"`
}

func NewGetAppLogsTokenReq() *GetAppLogsTokenReq {
	return &GetAppLogsTokenReq{}
}

func (req *GetAppLogsTokenReq) Validate() apperrors.ValidationErrors {
	var validators []vld.Validator
	validators = append(validators, basedto.ValidateID(&req.ProjectID, true, "projectId")...)
	validators = append(validators, basedto.ValidateID(&req.AppID, true, "appId")...)
	return apperrors.NewValidationErrors(vld.Validate(validators...))
}

type GetAppLogsTokenResp struct {
	Meta *basedto.Meta         `json:"meta"`
	Data *AppLogsTokenDataResp `json:"data"`
}

type AppLogsTokenDataResp struct {
	Token string `json:"token"`
}
