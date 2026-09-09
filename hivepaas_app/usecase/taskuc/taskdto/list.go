package taskdto

import (
	vld "github.com/tiendc/go-validator"

	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/basedto"
	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/entity/cacheentity"
	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/timeutil"
)

type ListTaskReq struct {
	Scope *entity.ObjectScope `json:"-" mapstructure:"-"`

	ScopeOnly    bool   `json:"-" mapstructure:"scopeOnly"`
	ProjectID    string `json:"-" mapstructure:"projectId"`
	ProjectEnvID string `json:"-" mapstructure:"projectEnvId"`
	AppID        string `json:"-" mapstructure:"appId"`

	Type     []base.TaskType   `json:"-" mapstructure:"type"`
	TargetID []string          `json:"-" mapstructure:"targetId"`
	Status   []base.TaskStatus `json:"-" mapstructure:"status"`
	FromDate timeutil.Date     `json:"-" mapstructure:"fromDate"`
	ToDate   timeutil.Date     `json:"-" mapstructure:"toDate"`
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

func (req *ListTaskReq) Validate() hperrors.ValidationErrors {
	validators := make([]vld.Validator, 0, 10) //nolint:mnd
	validators = append(validators, basedto.ValidateIDSlice(req.TargetID, true, 0, "targetId")...)
	validators = append(validators, basedto.ValidateSlice(req.Status, true, 0, base.AllTaskStatuses, "status")...)
	// TODO: add validation
	return hperrors.NewValidationErrors(vld.Validate(validators...))
}

type ListTaskResp struct {
	Meta *basedto.ListMeta `json:"meta"`
	Data []*TaskResp       `json:"data"`
}

func TransformTasks(
	tasks []*entity.Task,
	taskInfoMap map[string]*cacheentity.TaskInfo,
	refObjects *entity.RefObjects,
) (resp []*TaskResp, err error) {
	resp = make([]*TaskResp, 0, len(tasks))
	for _, task := range tasks {
		item, err := TransformTask(task, taskInfoMap[task.ID], refObjects)
		if err != nil {
			return nil, hperrors.Wrap(err)
		}
		resp = append(resp, item)
	}
	return resp, nil
}
