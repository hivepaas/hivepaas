package appdto

import (
	vld "github.com/tiendc/go-validator"

	"github.com/hivepaas/hivepaas/hivepaas_app/apperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/basedto"
	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
)

type ListAppReq struct {
	ProjectID string           `json:"-"`
	ParentID  string           `json:"-" mapstructure:"parentId"`
	Status    []base.AppStatus `json:"-" mapstructure:"status"`
	Env       []string         `json:"-" mapstructure:"env"`
	Search    string           `json:"-" mapstructure:"search"`
	GetStats  bool             `json:"-" mapstructure:"getStats"`

	Paging basedto.Paging `json:"-"`
}

func NewListAppReq() *ListAppReq {
	return &ListAppReq{
		Paging: basedto.Paging{
			// Default paging if unset by client
			Sort: basedto.Orders{{Direction: basedto.DirectionAsc, ColumnName: "created_at"}},
		},
	}
}

func (req *ListAppReq) Validate() apperrors.ValidationErrors {
	var validators []vld.Validator
	validators = append(validators, basedto.ValidateID(&req.ProjectID, true, "projectId")...)
	validators = append(validators, basedto.ValidateID(&req.ParentID, false, "parentId")...)
	validators = append(validators, basedto.ValidateSlice(req.Status, true, 0,
		base.AllAppStatuses, "status")...)
	return apperrors.NewValidationErrors(vld.Validate(validators...))
}

type ListAppResp struct {
	Meta *basedto.ListMeta `json:"meta"`
	Data []*AppResp        `json:"data"`
}

func TransformApps(apps []*entity.App, input *AppTransformationInput) ([]*AppResp, error) {
	resp := make([]*AppResp, 0, len(apps))
	for _, app := range apps {
		appResp, err := TransformApp(app, input)
		if err != nil {
			return nil, apperrors.New(err)
		}
		resp = append(resp, appResp)
	}
	return resp, nil
}
