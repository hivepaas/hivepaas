package appdto

import (
	vld "github.com/tiendc/go-validator"

	"github.com/localpaas/localpaas/localpaas_app/apperrors"
	"github.com/localpaas/localpaas/localpaas_app/base"
	"github.com/localpaas/localpaas/localpaas_app/basedto"
	"github.com/localpaas/localpaas/localpaas_app/entity"
)

type ListAppReq struct {
	ProjectID string           `json:"-"`
	Status    []base.AppStatus `json:"-" mapstructure:"status"`
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
	validators = append(validators, basedto.ValidateID(&req.ProjectID, true, "projectID")...)
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
			return nil, apperrors.Wrap(err)
		}
		resp = append(resp, appResp)
	}
	return resp, nil
}
