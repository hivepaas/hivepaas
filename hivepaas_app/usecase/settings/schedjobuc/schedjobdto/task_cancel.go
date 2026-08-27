package schedjobdto

import (
	vld "github.com/tiendc/go-validator"

	"github.com/hivepaas/hivepaas/hivepaas_app/basedto"
	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/settings"
)

type CancelSchedJobTaskReq struct {
	settings.BaseSettingReq
	JobID  string `json:"-"`
	TaskID string `json:"-"`
}

func NewCancelSchedJobTaskReq() *CancelSchedJobTaskReq {
	return &CancelSchedJobTaskReq{}
}

func (req *CancelSchedJobTaskReq) Validate() hperrors.ValidationErrors {
	validators := make([]vld.Validator, 0, 10) //nolint:mnd
	validators = append(validators, basedto.ValidateID(&req.JobID, true, "jobId")...)
	validators = append(validators, basedto.ValidateID(&req.TaskID, true, "taskId")...)
	return hperrors.NewValidationErrors(vld.Validate(validators...))
}

type CancelSchedJobTaskResp struct {
	Meta *basedto.Meta               `json:"meta"`
	Data *CancelSchedJobTaskDataResp `json:"data"`
}

type CancelSchedJobTaskDataResp struct {
	Canceled bool `json:"canceled"`
}
