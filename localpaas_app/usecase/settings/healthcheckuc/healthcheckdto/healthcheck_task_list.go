package healthcheckdto

import (
	vld "github.com/tiendc/go-validator"

	"github.com/localpaas/localpaas/localpaas_app/apperrors"
	"github.com/localpaas/localpaas/localpaas_app/base"
	"github.com/localpaas/localpaas/localpaas_app/basedto"
	"github.com/localpaas/localpaas/localpaas_app/usecase/settings"
	"github.com/localpaas/localpaas/localpaas_app/usecase/taskuc/taskdto"
)

type ListHealthcheckTaskReq struct {
	settings.BaseSettingReq
	JobID  string            `json:"-"`
	Status []base.TaskStatus `json:"-" mapstructure:"status"`
	Search string            `json:"-" mapstructure:"search"`
	Paging basedto.Paging    `json:"-"`
}

func NewListHealthcheckTaskReq() *ListHealthcheckTaskReq {
	return &ListHealthcheckTaskReq{
		Paging: basedto.Paging{
			// Default paging if unset by client
			Sort: basedto.Orders{{Direction: basedto.DirectionDesc, ColumnName: "created_at"}},
		},
	}
}

func (req *ListHealthcheckTaskReq) Validate() apperrors.ValidationErrors {
	var validators []vld.Validator
	validators = append(validators, basedto.ValidateID(&req.JobID, true, "jobID")...)
	validators = append(validators, basedto.ValidateSlice(req.Status, true, 0,
		base.AllTaskStatuses, "status")...)
	return apperrors.NewValidationErrors(vld.Validate(validators...))
}

type ListHealthcheckTaskResp struct {
	Meta *basedto.ListMeta   `json:"meta"`
	Data []*taskdto.TaskResp `json:"data"`
}
