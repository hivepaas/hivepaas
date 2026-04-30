package devhelperdto

import (
	"github.com/localpaas/localpaas/localpaas_app/apperrors"
	"github.com/localpaas/localpaas/localpaas_app/basedto"
	"github.com/localpaas/localpaas/localpaas_app/pkg/timeutil"
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
