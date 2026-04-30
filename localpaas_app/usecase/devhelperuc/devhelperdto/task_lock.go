package devhelperdto

import (
	"github.com/localpaas/localpaas/localpaas_app/apperrors"
	"github.com/localpaas/localpaas/localpaas_app/basedto"
	"github.com/localpaas/localpaas/localpaas_app/pkg/timeutil"
)

type LockTaskReq struct {
	TaskID   string            `json:"taskId"`
	Duration timeutil.Duration `json:"duration"`
}

func NewLockTaskReq() *LockTaskReq {
	return &LockTaskReq{}
}

func (req *LockTaskReq) Validate() apperrors.ValidationErrors {
	return nil
}

type LockTaskResp struct {
	Meta *basedto.Meta `json:"meta"`
}
