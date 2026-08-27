package hpappdto

import (
	"github.com/hivepaas/hivepaas/hivepaas_app/basedto"
	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
)

type ReloadHpAppConfigReq struct {
}

func NewReloadHpAppConfigReq() *ReloadHpAppConfigReq {
	return &ReloadHpAppConfigReq{}
}

// Validate implements interface basedto.ReqValidator
func (req *ReloadHpAppConfigReq) Validate() hperrors.ValidationErrors {
	return nil
}

type ReloadHpAppConfigResp struct {
	Meta *basedto.Meta `json:"meta"`
}
