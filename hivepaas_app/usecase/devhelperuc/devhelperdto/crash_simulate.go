package devhelperdto

import (
	"github.com/hivepaas/hivepaas/hivepaas_app/apperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/basedto"
)

type SimulateCrashReq struct {
}

func NewSimulateCrashReq() *SimulateCrashReq {
	return &SimulateCrashReq{}
}

func (req *SimulateCrashReq) Validate() apperrors.ValidationErrors {
	return nil
}

type SimulateCrashResp struct {
	Meta *basedto.Meta `json:"meta"`
}
