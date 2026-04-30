package lpappdto

import (
	"github.com/localpaas/localpaas/localpaas_app/apperrors"
	"github.com/localpaas/localpaas/localpaas_app/basedto"
)

type UpdateLpAppReq struct {
	TargetVersion string `json:"targetVersion"`
}

func NewUpdateLpAppReq() *UpdateLpAppReq {
	return &UpdateLpAppReq{}
}

// Validate implements interface basedto.ReqValidator
func (req *UpdateLpAppReq) Validate() apperrors.ValidationErrors {
	return nil
}

type UpdateLpAppResp struct {
	Meta *basedto.Meta `json:"meta"`
}
