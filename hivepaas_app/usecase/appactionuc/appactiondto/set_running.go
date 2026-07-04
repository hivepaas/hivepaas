package appactiondto

import (
	"github.com/hivepaas/hivepaas/hivepaas_app/apperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/basedto"
)

type SetAppRunningReq struct {
	ProjectID string `json:"-"`
	AppID     string `json:"-"`
	Running   bool   `json:"running"`
}

func NewSetAppRunningReq() *SetAppRunningReq {
	return &SetAppRunningReq{}
}

// Validate implements interface basedto.ReqValidator
func (req *SetAppRunningReq) Validate() apperrors.ValidationErrors {
	return nil
}

type SetAppRunningResp struct {
	Meta *basedto.Meta `json:"meta"`
}
