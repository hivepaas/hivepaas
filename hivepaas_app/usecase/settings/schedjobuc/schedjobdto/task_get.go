package schedjobdto

import (
	vld "github.com/tiendc/go-validator"

	"github.com/hivepaas/hivepaas/hivepaas_app/basedto"
	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
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

func (req *GetSchedJobTaskReq) Validate() hperrors.ValidationErrors {
	validators := make([]vld.Validator, 0, 10) //nolint:mnd
	validators = append(validators, basedto.ValidateID(&req.JobID, true, "jobId")...)
	validators = append(validators, basedto.ValidateID(&req.TaskID, true, "taskId")...)
	return hperrors.NewValidationErrors(vld.Validate(validators...))
}

type GetSchedJobTaskResp struct {
	Meta *basedto.Meta     `json:"meta"`
	Data *taskdto.TaskResp `json:"data"`
}
