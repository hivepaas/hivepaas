package apppreviewdto

import (
	vld "github.com/tiendc/go-validator"

	"github.com/hivepaas/hivepaas/hivepaas_app/apperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/basedto"
)

type PrepareCreatePreviewReq struct {
	ProjectID string `json:"-"`
	AppID     string `json:"-"`
}

func NewPrepareCreatePreviewReq() *PrepareCreatePreviewReq {
	return &PrepareCreatePreviewReq{}
}

func (req *PrepareCreatePreviewReq) Validate() apperrors.ValidationErrors {
	var validators []vld.Validator
	validators = append(validators, basedto.ValidateID(&req.ProjectID, true, "projectId")...)
	validators = append(validators, basedto.ValidateID(&req.AppID, true, "appId")...)
	return apperrors.NewValidationErrors(vld.Validate(validators...))
}

type PrepareCreatePreviewResp struct {
	Meta *basedto.Meta                 `json:"meta"`
	Data *PrepareCreatePreviewDataResp `json:"data"`
}

type PrepareCreatePreviewDataResp struct {
	RepoURL             string                `json:"repoURL"`
	RepoCredentials     *basedto.ObjectIDResp `json:"repoCredentials"`
	CanListBranches     bool                  `json:"canListBranches"`
	CanListPullRequests bool                  `json:"canListPullRequests"`
}
