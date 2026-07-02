package schedjobdto

import (
	vld "github.com/tiendc/go-validator"

	"github.com/hivepaas/hivepaas/hivepaas_app/apperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/basedto"
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

func (req *CancelSchedJobTaskReq) Validate() apperrors.ValidationErrors {
	var validators []vld.Validator
	validators = append(validators, basedto.ValidateID(&req.JobID, true, "jobId")...)
	validators = append(validators, basedto.ValidateID(&req.TaskID, true, "taskId")...)
	return apperrors.NewValidationErrors(vld.Validate(validators...))
}

type CancelSchedJobTaskResp struct {
	Meta *basedto.Meta               `json:"meta"`
	Data *CancelSchedJobTaskDataResp `json:"data"`
}

type CancelSchedJobTaskDataResp struct {
	Canceled bool `json:"canceled"`
}
