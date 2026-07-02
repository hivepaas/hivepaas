package hpappdto

import (
	"github.com/hivepaas/hivepaas/hivepaas_app/apperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/basedto"
)

type UpdateHpAppReq struct {
	TargetVersion string `json:"targetVersion"`
}

func NewUpdateHpAppReq() *UpdateHpAppReq {
	return &UpdateHpAppReq{}
}

// Validate implements interface basedto.ReqValidator
func (req *UpdateHpAppReq) Validate() apperrors.ValidationErrors {
	return nil
}

type UpdateHpAppResp struct {
	Meta *basedto.Meta `json:"meta"`
}
