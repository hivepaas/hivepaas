package appdto

import (
	vld "github.com/tiendc/go-validator"

	"github.com/hivepaas/hivepaas/hivepaas_app/basedto"
	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
)

type GetAppLogsInfoReq struct {
	ProjectID    string `json:"-"`
	ProjectEnvID string `json:"-"`
	AppID        string `json:"-"`
}

func NewGetAppLogsInfoReq() *GetAppLogsInfoReq {
	return &GetAppLogsInfoReq{}
}

func (req *GetAppLogsInfoReq) Validate() hperrors.ValidationErrors {
	validators := make([]vld.Validator, 0, 10) //nolint:mnd
	validators = append(validators, basedto.ValidateID(&req.ProjectID, true, "projectId")...)
	validators = append(validators, basedto.ValidateID(&req.ProjectEnvID, true, "projectEnv")...)
	validators = append(validators, basedto.ValidateID(&req.AppID, true, "appId")...)
	return hperrors.NewValidationErrors(vld.Validate(validators...))
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
