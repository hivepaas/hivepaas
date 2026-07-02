package appactiondto

import (
	"github.com/hivepaas/hivepaas/hivepaas_app/apperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/basedto"
)

type RestartAppReq struct {
	ProjectID string `json:"-"`
	AppID     string `json:"-"`
}

func NewRestartAppReq() *RestartAppReq {
	return &RestartAppReq{}
}

// Validate implements interface basedto.ReqValidator
func (req *RestartAppReq) Validate() apperrors.ValidationErrors {
	return nil
}

type RestartAppResp struct {
	Meta *basedto.Meta `json:"meta"`
}
