package hpappdto

import (
	"github.com/hivepaas/hivepaas/hivepaas_app/apperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/basedto"
)

type RestartHpAppReq struct {
	RestartMainApp  bool `json:"restartMainApp"`
	RestartDbApp    bool `json:"restartDbApp"`
	RestartCacheApp bool `json:"restartCacheApp"`
}

func NewRestartHpAppReq() *RestartHpAppReq {
	return &RestartHpAppReq{}
}

// Validate implements interface basedto.ReqValidator
func (req *RestartHpAppReq) Validate() apperrors.ValidationErrors {
	return nil
}

type RestartHpAppResp struct {
	Meta *basedto.Meta `json:"meta"`
}
