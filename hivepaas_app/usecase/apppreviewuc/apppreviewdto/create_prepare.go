package apppreviewdto

import (
	vld "github.com/tiendc/go-validator"

	"github.com/hivepaas/hivepaas/hivepaas_app/basedto"
	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
)

type PrepareCreatePreviewReq struct {
	ProjectID    string `json:"-"`
	ProjectEnvID string `json:"-"`
	AppID        string `json:"-"`
}

func NewPrepareCreatePreviewReq() *PrepareCreatePreviewReq {
	return &PrepareCreatePreviewReq{}
}

func (req *PrepareCreatePreviewReq) Validate() hperrors.ValidationErrors {
	validators := make([]vld.Validator, 0, 10) //nolint:mnd
	validators = append(validators, basedto.ValidateID(&req.ProjectID, true, "projectId")...)
	validators = append(validators, basedto.ValidateID(&req.ProjectEnvID, true, "projectEnv")...)
	validators = append(validators, basedto.ValidateID(&req.AppID, true, "appId")...)
	return hperrors.NewValidationErrors(vld.Validate(validators...))
}

type PrepareCreatePreviewResp struct {
	Meta *basedto.Meta                 `json:"meta"`
	Data *PrepareCreatePreviewDataResp `json:"data"`
}

type PrepareCreatePreviewDataResp struct {
	Enabled              bool                  `json:"enabled"`
	RepoURL              string                `json:"repoURL"`
	RepoCredentials      *basedto.ObjectIDResp `json:"repoCredentials"`
	CanListBranches      bool                  `json:"canListBranches"`
	CanListPullRequests  bool                  `json:"canListPullRequests"`
	CanCloneDBApps       bool                  `json:"canCloneDbApps"`
	CanSkipCloningDBApps bool                  `json:"canSkipCloningDbApps"`
}
