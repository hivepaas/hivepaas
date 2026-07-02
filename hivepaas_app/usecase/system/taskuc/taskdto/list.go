package taskdto

import (
	vld "github.com/tiendc/go-validator"

	"github.com/hivepaas/hivepaas/hivepaas_app/apperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/basedto"
	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/entity/cacheentity"
)

type ListTaskReq struct {
	JobName  base.SystemJobName `json:"-" mapstructure:"jobName"`
	TargetID []string           `json:"-" mapstructure:"targetId"`
	Status   []base.TaskStatus  `json:"-" mapstructure:"status"`
	Search   string             `json:"-" mapstructure:"search"`

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
	validators = append(validators, basedto.ValidateStrIn(&req.JobName, false, base.AllSystemJobNames, "jobName")...)
	validators = append(validators, basedto.ValidateIDSlice(req.TargetID, true, 0, "targetId")...)
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
			return nil, apperrors.New(err)
		}
		resp = append(resp, item)
	}
	return resp, nil
}
