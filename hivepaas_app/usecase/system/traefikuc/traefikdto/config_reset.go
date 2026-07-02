package traefikdto

import (
	"github.com/hivepaas/hivepaas/hivepaas_app/apperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/basedto"
)

type ResetTraefikConfigReq struct {
}

func NewResetTraefikConfigReq() *ResetTraefikConfigReq {
	return &ResetTraefikConfigReq{}
}

// Validate implements interface basedto.ReqValidator
func (req *ResetTraefikConfigReq) Validate() apperrors.ValidationErrors {
	return nil
}

type ResetTraefikConfigResp struct {
	Meta *basedto.Meta `json:"meta"`
}
