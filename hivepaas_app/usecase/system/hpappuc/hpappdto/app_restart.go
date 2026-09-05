package hpappdto

import (
	"github.com/hivepaas/hivepaas/hivepaas_app/basedto"
	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
)

type RestartHpAppReq struct {
	RestartMainApp  bool `json:"restartMainApp"`
	RestartWorkers  bool `json:"restartWorkers"`
	RestartAgents   bool `json:"restartAgents"`
	RestartDbApp    bool `json:"restartDbApp"`
	RestartCacheApp bool `json:"restartCacheApp"`
}

func NewRestartHpAppReq() *RestartHpAppReq {
	return &RestartHpAppReq{}
}

// Validate implements interface basedto.ReqValidator
func (req *RestartHpAppReq) Validate() hperrors.ValidationErrors {
	return nil
}

type RestartHpAppResp struct {
	Meta *basedto.Meta `json:"meta"`
}
