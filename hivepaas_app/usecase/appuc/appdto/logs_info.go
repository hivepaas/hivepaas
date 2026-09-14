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
	// History says whether stored logs can be shown, and if not why.
	History *AppLogHistoryInfoResp `json:"history"`
}

type AppLogHistoryInfoResp struct {
	Available bool   `json:"available"`
	Reason    string `json:"reason,omitempty"`
	// Retention is how far back stored logs reach, written the way a duration
	// is written everywhere else - "30d", "12h", "1d12h". It is empty when the
	// backend is not one HivePaaS keeps, and the caller then knows nothing
	// about the depth rather than assuming a default.
	Retention string `json:"retention,omitempty"`
}

type TaskLogsInfoResp struct {
	ID string `json:"id"`
}
