package traefikdto

import (
	"github.com/hivepaas/hivepaas/hivepaas_app/basedto"
	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
)

type ReloadTraefikConfigReq struct {
}

func NewReloadTraefikConfigReq() *ReloadTraefikConfigReq {
	return &ReloadTraefikConfigReq{}
}

// Validate implements interface basedto.ReqValidator
func (req *ReloadTraefikConfigReq) Validate() hperrors.ValidationErrors {
	return nil
}

type ReloadTraefikConfigResp struct {
	Meta *basedto.Meta `json:"meta"`
}
