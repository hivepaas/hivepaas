package apppreviewdto

import (
	vld "github.com/tiendc/go-validator"

	"github.com/hivepaas/hivepaas/hivepaas_app/apperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/basedto"
)

const (
	repoRefMaxLen   = 100
	subdomainMaxLen = 30
)

type CreatePreviewReq struct {
	ProjectID string `json:"-"`
	AppID     string `json:"-"`

	RepoRef         string `json:"repoRef"`
	CustomSubdomain string `json:"customSubdomain"`
	NoStart         bool   `json:"noStart"`
}

func NewCreatePreviewReq() *CreatePreviewReq {
	return &CreatePreviewReq{}
}

func (req *CreatePreviewReq) Validate() apperrors.ValidationErrors {
	var validators []vld.Validator
	validators = append(validators, basedto.ValidateID(&req.ProjectID, true, "projectId")...)
	validators = append(validators, basedto.ValidateID(&req.AppID, true, "appId")...)
	validators = append(validators, basedto.ValidateStr(&req.RepoRef, true, 1, repoRefMaxLen,
		"repoRef")...)
	validators = append(validators, basedto.ValidateStr(&req.CustomSubdomain, false, 1, subdomainMaxLen,
		"customSubdomain")...)
	return apperrors.NewValidationErrors(vld.Validate(validators...))
}

type CreatePreviewResp struct {
	Meta *basedto.Meta         `json:"meta"`
	Data *basedto.ObjectIDResp `json:"data"`
}
