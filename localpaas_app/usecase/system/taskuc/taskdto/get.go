package taskdto

import (
	"time"

	vld "github.com/tiendc/go-validator"

	"github.com/localpaas/localpaas/localpaas_app/apperrors"
	"github.com/localpaas/localpaas/localpaas_app/base"
	"github.com/localpaas/localpaas/localpaas_app/basedto"
	"github.com/localpaas/localpaas/localpaas_app/entity"
	"github.com/localpaas/localpaas/localpaas_app/entity/cacheentity"
	"github.com/localpaas/localpaas/localpaas_app/pkg/copier"
)

type GetTaskReq struct {
	ID string `json:"-"`
}

func NewGetTaskReq() *GetTaskReq {
	return &GetTaskReq{}
}

func (req *GetTaskReq) Validate() apperrors.ValidationErrors {
	var validators []vld.Validator
	validators = append(validators, basedto.ValidateID(&req.ID, true, "id")...)
	return apperrors.NewValidationErrors(vld.Validate(validators...))
}

type GetTaskResp struct {
	Meta *basedto.Meta `json:"meta"`
	Data *TaskResp     `json:"data"`
}

type TaskResp struct {
	ID        string            `json:"id"`
	Type      base.TaskType     `json:"type"`
	Status    base.TaskStatus   `json:"status"`
	Config    entity.TaskConfig `json:"config"` // NOTE: use entity's type directly here, may need refactor
	LastError string            `json:"lastError"`
	UpdateVer int               `json:"updateVer"`

	RunAt     *time.Time `json:"runAt" copy:",nilonzero"`
	RetryAt   *time.Time `json:"retryAt" copy:",nilonzero"`
	StartedAt *time.Time `json:"startedAt" copy:",nilonzero"`
	EndedAt   *time.Time `json:"endedAt" copy:",nilonzero"`

	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`
}

func TransformTask(task *entity.Task, taskInfo *cacheentity.TaskInfo) (resp *TaskResp, err error) {
	if err = copier.Copy(&resp, &task); err != nil {
		return nil, apperrors.Wrap(err)
	}
	if taskInfo != nil {
		resp.Status = taskInfo.Status
		if taskInfo.Status == base.TaskStatusInProgress {
			resp.StartedAt = &taskInfo.StartedAt
		}
	}

	if resp.Status == base.TaskStatusInProgress || resp.Status == base.TaskStatusFailed {
		runs, _ := task.GetRuns()
		if len(runs) > 0 {
			resp.LastError = runs[len(runs)-1].Error
		}
	}

	return resp, nil
}
