package taskdto

import (
	vld "github.com/tiendc/go-validator"

	"github.com/localpaas/localpaas/localpaas_app/apperrors"
	"github.com/localpaas/localpaas/localpaas_app/base"
	"github.com/localpaas/localpaas/localpaas_app/basedto"
	"github.com/localpaas/localpaas/localpaas_app/entity"
	"github.com/localpaas/localpaas/localpaas_app/entity/cacheentity"
)

type ListTaskReq struct {
	TargetID []string          `json:"-" mapstructure:"targetID"`
	Status   []base.TaskStatus `json:"-" mapstructure:"status"`
	Search   string            `json:"-" mapstructure:"search"`

	Paging basedto.Paging `json:"-"`
}

func NewListTaskReq() *ListTaskReq {
	return &ListTaskReq{
		Paging: basedto.Paging{
			// Default paging if unset by client
			Sort: basedto.Orders{{Direction: basedto.DirectionDesc, ColumnName: "created_at"}},
		},
	}
}

func (req *ListTaskReq) Validate() apperrors.ValidationErrors {
	var validators []vld.Validator
	validators = append(validators, basedto.ValidateIDSlice(req.TargetID, true, 0, "targetID")...)
	validators = append(validators, basedto.ValidateSlice(req.Status, true, 0, base.AllTaskStatuses, "status")...)
	return apperrors.NewValidationErrors(vld.Validate(validators...))
}

type ListTaskResp struct {
	Meta *basedto.ListMeta `json:"meta"`
	Data []*TaskResp       `json:"data"`
}

func TransformTasks(tasks []*entity.Task, taskInfoMap map[string]*cacheentity.TaskInfo) (resp []*TaskResp, err error) {
	resp = make([]*TaskResp, 0, len(tasks))
	for _, task := range tasks {
		item, err := TransformTask(task, taskInfoMap[task.ID])
		if err != nil {
			return nil, apperrors.Wrap(err)
		}
		resp = append(resp, item)
	}
	return resp, nil
}
