package syserrordto

import (
	"github.com/hivepaas/hivepaas/hivepaas_app/basedto"
	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
)

type CreateSysErrorReq struct {
	ErrorInfo *hperrors.ErrorInfo `json:"-"`
}

func NewCreateSysErrorReq() *CreateSysErrorReq {
	return &CreateSysErrorReq{}
}

// Validate implements interface basedto.ReqValidator
func (req *CreateSysErrorReq) Validate() hperrors.ValidationErrors {
	return nil
}

type CreateSysErrorResp struct {
	Meta *basedto.Meta         `json:"meta"`
	Data *basedto.ObjectIDResp `json:"data"`
}
