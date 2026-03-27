package cronjobdto

import (
	vld "github.com/tiendc/go-validator"

	"github.com/localpaas/localpaas/localpaas_app/apperrors"
	"github.com/localpaas/localpaas/localpaas_app/basedto"
	"github.com/localpaas/localpaas/localpaas_app/usecase/settings"
	"github.com/localpaas/localpaas/localpaas_app/usecase/taskuc/taskdto"
)

type GetCronJobTaskReq struct {
	settings.BaseSettingReq
	JobID  string `json:"-"`
	TaskID string `json:"-"`
}

func NewGetCronJobTaskReq() *GetCronJobTaskReq {
	return &GetCronJobTaskReq{}
}

func (req *GetCronJobTaskReq) Validate() apperrors.ValidationErrors {
	var validators []vld.Validator
	validators = append(validators, basedto.ValidateID(&req.JobID, true, "jobID")...)
	validators = append(validators, basedto.ValidateID(&req.TaskID, true, "taskID")...)
	return apperrors.NewValidationErrors(vld.Validate(validators...))
}

type GetCronJobTaskResp struct {
	Meta *basedto.Meta     `json:"meta"`
	Data *taskdto.TaskResp `json:"data"`
}
