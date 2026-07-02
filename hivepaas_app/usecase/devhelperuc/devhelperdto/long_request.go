package devhelperdto

import (
	"github.com/hivepaas/hivepaas/hivepaas_app/apperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/basedto"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/timeutil"
)

type LongRequestReq struct {
	Duration timeutil.Duration `json:"duration"`
}

func NewLongRequestReq() *LongRequestReq {
	return &LongRequestReq{}
}

func (req *LongRequestReq) Validate() apperrors.ValidationErrors {
	return nil
}

type LongRequestResp struct {
	Meta *basedto.Meta `json:"meta"`
}
