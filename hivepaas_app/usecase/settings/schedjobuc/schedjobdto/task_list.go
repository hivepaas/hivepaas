package schedjobdto

import (
	vld "github.com/tiendc/go-validator"

	"github.com/hivepaas/hivepaas/hivepaas_app/apperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/basedto"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/settings"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/system/taskuc/taskdto"
)

type ListSchedJobTaskReq struct {
	settings.BaseSettingReq
	JobID  string            `json:"-"`
	Status []base.TaskStatus `json:"-" mapstructure:"status"`
	Search string            `json:"-" mapstructure:"search"`
	Paging basedto.Paging    `json:"-"`
}

func NewListSchedJobTaskReq() *ListSchedJobTaskReq {
	return &ListSchedJobTaskReq{
		Paging: basedto.Paging{
			// Default paging if unset by client
			Sort: basedto.Orders{{Direction: basedto.DirectionDesc, ColumnName: "created_at"}},
		},
	}
}

func (req *ListSchedJobTaskReq) Validate() apperrors.ValidationErrors {
	var validators []vld.Validator
	validators = append(validators, basedto.ValidateID(&req.JobID, true, "jobId")...)
	validators = append(validators, basedto.ValidateSlice(req.Status, true, 0,
		base.AllTaskStatuses, "status")...)
	return apperrors.NewValidationErrors(vld.Validate(validators...))
}

type ListSchedJobTaskResp struct {
	Meta *basedto.ListMeta   `json:"meta"`
	Data []*taskdto.TaskResp `json:"data"`
}
