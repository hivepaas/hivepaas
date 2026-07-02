package hpappdto

import (
	"github.com/hivepaas/hivepaas/hivepaas_app/apperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/basedto"
)

type ReloadHpAppConfigReq struct {
}

func NewReloadHpAppConfigReq() *ReloadHpAppConfigReq {
	return &ReloadHpAppConfigReq{}
}

// Validate implements interface basedto.ReqValidator
func (req *ReloadHpAppConfigReq) Validate() apperrors.ValidationErrors {
	return nil
}

type ReloadHpAppConfigResp struct {
	Meta *basedto.Meta `json:"meta"`
}
