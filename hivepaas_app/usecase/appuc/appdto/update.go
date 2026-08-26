package appdto

import (
	vld "github.com/tiendc/go-validator"
	"github.com/tiendc/gofn"

	"github.com/hivepaas/hivepaas/hivepaas_app/apperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/basedto"
)

type UpdateAppReq struct {
	ProjectID    string `json:"-"`
	ProjectEnvID string `json:"-"`
	AppID        string `json:"-"`
	UpdateVer    int    `json:"updateVer"`
	*AppBaseReq
}

func NewUpdateAppReq() *UpdateAppReq {
	return &UpdateAppReq{}
}

// Validate implements interface basedto.ReqValidator
func (req *UpdateAppReq) Validate() apperrors.ValidationErrors {
	validators := make([]vld.Validator, 0, 10) //nolint:mnd
	validators = append(validators, basedto.ValidateID(&req.ProjectID, true, "projectId")...)
	validators = append(validators, basedto.ValidateID(&req.ProjectEnvID, true, "projectEnv")...)
	validators = append(validators, basedto.ValidateID(&req.AppID, true, "appId")...)
	validators = append(validators, req.validate("")...)
	return apperrors.NewValidationErrors(vld.Validate(validators...))
}

func (req *UpdateAppReq) ModifyRequest() error {
	// We don't allow updating `status` (just assign a value if unset, UC will ignore it)
	req.Status = gofn.Coalesce(req.Status, base.AppStatusActive)
	return req.modifyRequest()
}

type UpdateAppResp struct {
	Meta *basedto.Meta `json:"meta"`
}
