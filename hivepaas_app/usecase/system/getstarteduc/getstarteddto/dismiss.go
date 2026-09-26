package getstarteddto

import (
	"github.com/hivepaas/hivepaas/hivepaas_app/basedto"
)

type DismissReq struct{}

func NewDismissReq() *DismissReq {
	return &DismissReq{}
}

type DismissResp struct {
	Meta *basedto.Meta `json:"meta"`
}
