package projectdto

import (
	vld "github.com/tiendc/go-validator"
	"github.com/tiendc/gofn"

	"github.com/hivepaas/hivepaas/hivepaas_app/apperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/basedto"
)

type UpdateProjectReq struct {
	ID        string `json:"-"`
	UpdateVer int    `json:"updateVer"`
	*ProjectBaseReq
}

func NewUpdateProjectReq() *UpdateProjectReq {
	return &UpdateProjectReq{}
}

// Validate implements interface basedto.ReqValidator
func (req *UpdateProjectReq) Validate() apperrors.ValidationErrors {
	validators := make([]vld.Validator, 0, 5) //nolint:mnd
	validators = append(validators, req.validate("")...)
	return apperrors.NewValidationErrors(vld.Validate(validators...))
}

func (req *UpdateProjectReq) ModifyRequest() error {
	// We don't allow updating `status` (just assign a value if unset, UC will ignore it)
	req.Status = gofn.Coalesce(req.Status, base.ProjectStatusActive)
	return req.modifyRequest()
}

type UpdateProjectResp struct {
	Meta *basedto.Meta `json:"meta"`
}
