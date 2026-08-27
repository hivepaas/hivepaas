package taskdto

import (
	vld "github.com/tiendc/go-validator"

	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/basedto"
	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/entity/cacheentity"
	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
)

type GetTaskStatusReq struct {
	ID string `json:"-"`
}

func NewGetTaskStatusReq() *GetTaskStatusReq {
	return &GetTaskStatusReq{}
}

func (req *GetTaskStatusReq) Validate() hperrors.ValidationErrors {
	validators := make([]vld.Validator, 0, 10) //nolint:mnd
	validators = append(validators, basedto.ValidateID(&req.ID, true, "id")...)
	return hperrors.NewValidationErrors(vld.Validate(validators...))
}

type GetTaskStatusResp struct {
	Meta *basedto.Meta   `json:"meta"`
	Data *TaskStatusResp `json:"data"`
}

type TaskStatusResp struct {
	Status base.TaskStatus `json:"status"`
}

func TransformTaskStatus(task *entity.Task, taskInfo *cacheentity.TaskInfo) *TaskStatusResp {
	resp := &TaskStatusResp{
		Status: task.Status,
	}
	if taskInfo != nil {
		resp.Status = taskInfo.Status
	}
	return resp
}
