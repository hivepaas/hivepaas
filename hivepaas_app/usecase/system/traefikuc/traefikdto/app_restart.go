package traefikdto

import (
	"github.com/hivepaas/hivepaas/hivepaas_app/basedto"
	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
)

type RestartTraefikReq struct {
}

func NewRestartTraefikReq() *RestartTraefikReq {
	return &RestartTraefikReq{}
}

// Validate implements interface basedto.ReqValidator
func (req *RestartTraefikReq) Validate() hperrors.ValidationErrors {
	return nil
}

type RestartTraefikResp struct {
	Meta *basedto.Meta `json:"meta"`
}
