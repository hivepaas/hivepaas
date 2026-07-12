package basicauthdto

import (
	vld "github.com/tiendc/go-validator"

	"github.com/hivepaas/hivepaas/hivepaas_app/apperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/basedto"
	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/settings"
)

type ListBasicAuthReq struct {
	settings.ListSettingReq
}

func NewListBasicAuthReq() *ListBasicAuthReq {
	return &ListBasicAuthReq{
		ListSettingReq: settings.ListSettingReq{
			Paging: basedto.Paging{
				// Default paging if unset by client
				Sort: basedto.Orders{{Direction: basedto.DirectionAsc, ColumnName: "name"}},
			},
		},
	}
}

func (req *ListBasicAuthReq) Validate() apperrors.ValidationErrors {
	var validators []vld.Validator
	validators = append(validators, req.ListSettingReq.Validate()...)
	return apperrors.NewValidationErrors(vld.Validate(validators...))
}

type ListBasicAuthResp struct {
	Meta *basedto.ListMeta `json:"meta"`
	Data []*BasicAuthResp  `json:"data"`
}

func TransformBasicAuths(
	settings []*entity.Setting,
	refObjects *entity.RefObjects,
) (resp []*BasicAuthResp, err error) {
	resp = make([]*BasicAuthResp, 0, len(settings))
	for _, setting := range settings {
		item, err := TransformBasicAuth(setting, refObjects)
		if err != nil {
			return nil, apperrors.Wrap(err)
		}
		resp = append(resp, item)
	}
	return resp, nil
}
