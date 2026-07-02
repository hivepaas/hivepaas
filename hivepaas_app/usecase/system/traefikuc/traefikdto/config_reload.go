package traefikdto

import (
	"github.com/hivepaas/hivepaas/hivepaas_app/apperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/basedto"
)

type ReloadTraefikConfigReq struct {
}

func NewReloadTraefikConfigReq() *ReloadTraefikConfigReq {
	return &ReloadTraefikConfigReq{}
}

// Validate implements interface basedto.ReqValidator
func (req *ReloadTraefikConfigReq) Validate() apperrors.ValidationErrors {
	return nil
}

type ReloadTraefikConfigResp struct {
	Meta *basedto.Meta `json:"meta"`
}
