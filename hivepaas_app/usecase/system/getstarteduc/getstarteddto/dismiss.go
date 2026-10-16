package getstarteddto

import (
	"github.com/hivepaas/hivepaas/hivepaas_app/basedto"
	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
)

type DismissReq struct{}

func NewDismissReq() *DismissReq {
	return &DismissReq{}
}

// Validate implements interface basedto.ReqValidator
func (req *DismissReq) Validate() hperrors.ValidationErrors {
	return nil
}

type DismissResp struct {
	Meta *basedto.Meta `json:"meta"`
}
