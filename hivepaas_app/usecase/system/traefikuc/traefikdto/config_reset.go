package traefikdto

import (
	"github.com/hivepaas/hivepaas/hivepaas_app/basedto"
	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
)

type ResetTraefikConfigReq struct {
}

func NewResetTraefikConfigReq() *ResetTraefikConfigReq {
	return &ResetTraefikConfigReq{}
}

// Validate implements interface basedto.ReqValidator
func (req *ResetTraefikConfigReq) Validate() hperrors.ValidationErrors {
	return nil
}

type ResetTraefikConfigResp struct {
	Meta *basedto.Meta `json:"meta"`
}
