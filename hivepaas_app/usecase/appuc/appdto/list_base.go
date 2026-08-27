package appdto

import (
	vld "github.com/tiendc/go-validator"

	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/basedto"
	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
)

type ListAppBaseReq struct {
	ProjectID    string           `json:"-"`
	ProjectEnvID string           `json:"-"`
	ParentID     string           `json:"-" mapstructure:"parentId"`
	Status       []base.AppStatus `json:"-" mapstructure:"status"`
	Env          string           `json:"-" mapstructure:"env"`
	Search       string           `json:"-" mapstructure:"search"`

	Paging basedto.Paging `json:"-"`
}

func NewListAppBaseReq() *ListAppBaseReq {
	return &ListAppBaseReq{
		Status: []base.AppStatus{base.AppStatusActive},
		Paging: basedto.Paging{
			// Default paging if unset by client
			Sort: basedto.Orders{{Direction: basedto.DirectionAsc, ColumnName: "name"}},
		},
	}
}

// Validate implements interface basedto.ReqValidator
func (req *ListAppBaseReq) Validate() hperrors.ValidationErrors {
	validators := make([]vld.Validator, 0, 10) //nolint:mnd
	validators = append(validators, basedto.ValidateID(&req.ProjectID, true, "projectId")...)
	validators = append(validators, basedto.ValidateID(&req.ProjectEnvID, false, "projectEnv")...)
	validators = append(validators, basedto.ValidateID(&req.ParentID, false, "parentId")...)
	validators = append(validators, basedto.ValidateSlice(req.Status, true, 0,
		base.AllAppStatuses, "status")...)
	return hperrors.NewValidationErrors(vld.Validate(validators...))
}

type ListAppBaseResp struct {
	Meta *basedto.ListMeta `json:"meta"`
	Data []*AppBaseResp    `json:"data"`
}
