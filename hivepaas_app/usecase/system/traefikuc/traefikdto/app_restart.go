package traefikdto

import (
	"github.com/hivepaas/hivepaas/hivepaas_app/apperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/basedto"
)

type RestartTraefikReq struct {
}

func NewRestartTraefikReq() *RestartTraefikReq {
	return &RestartTraefikReq{}
}

// Validate implements interface basedto.ReqValidator
func (req *RestartTraefikReq) Validate() apperrors.ValidationErrors {
	return nil
}

type RestartTraefikResp struct {
	Meta *basedto.Meta `json:"meta"`
}
