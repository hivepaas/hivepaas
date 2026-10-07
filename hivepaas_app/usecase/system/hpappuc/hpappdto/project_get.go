package hpappdto

import (
	"github.com/hivepaas/hivepaas/hivepaas_app/basedto"
	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
)

type GetHpProjectReq struct {
}

func NewGetHpProjectReq() *GetHpProjectReq {
	return &GetHpProjectReq{}
}

// Validate implements interface basedto.ReqValidator
func (req *GetHpProjectReq) Validate() hperrors.ValidationErrors {
	return nil
}

type GetHpProjectResp struct {
	Meta *basedto.Meta  `json:"meta"`
	Data *HpProjectResp `json:"data"`
}

// HpProjectResp is the project HivePaaS runs in: its own apps, and those it
// provisions. The projects list leaves it out.
type HpProjectResp struct {
	ID   string `json:"id"`
	Name string `json:"name"`
	Key  string `json:"key"`
}
