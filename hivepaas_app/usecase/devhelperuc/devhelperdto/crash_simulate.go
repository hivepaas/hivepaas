package devhelperdto

import (
	"github.com/hivepaas/hivepaas/hivepaas_app/basedto"
	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
)

type SimulateCrashReq struct {
}

func NewSimulateCrashReq() *SimulateCrashReq {
	return &SimulateCrashReq{}
}

func (req *SimulateCrashReq) Validate() hperrors.ValidationErrors {
	return nil
}

type SimulateCrashResp struct {
	Meta *basedto.Meta `json:"meta"`
}
