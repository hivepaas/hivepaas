package schedjobdto

import (
	vld "github.com/tiendc/go-validator"

	"github.com/localpaas/localpaas/localpaas_app/apperrors"
	"github.com/localpaas/localpaas/localpaas_app/basedto"
	"github.com/localpaas/localpaas/localpaas_app/usecase/settings"
	"github.com/localpaas/localpaas/localpaas_app/usecase/taskuc/taskdto"
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
