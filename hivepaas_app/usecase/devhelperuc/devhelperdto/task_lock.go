package devhelperdto

import (
	"github.com/hivepaas/hivepaas/hivepaas_app/basedto"
	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/timeutil"
)

type LockTaskReq struct {
	TaskID   string            `json:"taskId"`
	Duration timeutil.Duration `json:"duration"`
}

func NewLockTaskReq() *LockTaskReq {
	return &LockTaskReq{}
}

func (req *LockTaskReq) Validate() hperrors.ValidationErrors {
	return nil
}

type LockTaskResp struct {
	Meta *basedto.Meta `json:"meta"`
}
