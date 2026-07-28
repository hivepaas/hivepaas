package appdto

import (
	vld "github.com/tiendc/go-validator"

	"github.com/hivepaas/hivepaas/hivepaas_app/apperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/basedto"
)

type GetAppLogsInfoReq struct {
	ProjectID    string `json:"-"`
	ProjectEnvID string `json:"-"`
	AppID        string `json:"-"`
}

func NewGetAppLogsInfoReq() *GetAppLogsInfoReq {
	return &GetAppLogsInfoReq{}
}

func (req *GetAppLogsInfoReq) Validate() apperrors.ValidationErrors {
	var validators []vld.Validator
	validators = append(validators, basedto.ValidateID(&req.ProjectID, true, "projectId")...)
	validators = append(validators, basedto.ValidateStr(&req.ProjectEnvID, true,
		base.ProjectEnvMinLen, base.ProjectEnvMaxLen, "projectEnv")...)
	validators = append(validators, basedto.ValidateID(&req.AppID, true, "appId")...)
	return apperrors.NewValidationErrors(vld.Validate(validators...))
}

type GetAppLogsInfoResp struct {
	Meta *basedto.Meta        `json:"meta"`
	Data *AppLogsInfoDataResp `json:"data"`
}

type AppLogsInfoDataResp struct {
	Enabled bool                `json:"enabled"`
	Tasks   []*TaskLogsInfoResp `json:"tasks"`
}

type TaskLogsInfoResp struct {
	ID string `json:"id"`
}
