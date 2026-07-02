package schedjobdto

import (
	vld "github.com/tiendc/go-validator"

	"github.com/hivepaas/hivepaas/hivepaas_app/apperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/basedto"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/settings"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/system/taskuc/taskdto"
)

type GetSchedJobTaskReq struct {
	settings.BaseSettingReq
	JobID  string `json:"-"`
	TaskID string `json:"-"`
}

func NewGetSchedJobTaskReq() *GetSchedJobTaskReq {
	return &GetSchedJobTaskReq{}
}

func (req *GetSchedJobTaskReq) Validate() apperrors.ValidationErrors {
	var validators []vld.Validator
	validators = append(validators, basedto.ValidateID(&req.JobID, true, "jobId")...)
	validators = append(validators, basedto.ValidateID(&req.TaskID, true, "taskId")...)
	return apperrors.NewValidationErrors(vld.Validate(validators...))
}

type GetSchedJobTaskResp struct {
	Meta *basedto.Meta     `json:"meta"`
	Data *taskdto.TaskResp `json:"data"`
}
