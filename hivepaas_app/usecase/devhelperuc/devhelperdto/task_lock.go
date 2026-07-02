package devhelperdto

import (
	"github.com/hivepaas/hivepaas/hivepaas_app/apperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/basedto"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/timeutil"
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
