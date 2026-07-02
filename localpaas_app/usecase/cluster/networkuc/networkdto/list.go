package networkdto

import (
	vld "github.com/tiendc/go-validator"

	"github.com/localpaas/localpaas/localpaas_app/apperrors"
	"github.com/localpaas/localpaas/localpaas_app/basedto"
	"github.com/localpaas/localpaas/localpaas_app/entity"
	"github.com/localpaas/localpaas/localpaas_app/usecase/settings"
)

type ListNetworkReq struct {
	settings.ListSettingReq
}

func NewListNetworkReq() *ListNetworkReq {
	return &ListNetworkReq{
		ListSettingReq: settings.ListSettingReq{
			Paging: basedto.Paging{
				// Default paging if unset by client
				Sort: basedto.Orders{{Direction: basedto.DirectionAsc, ColumnName: "name"}},
			},
		},
	}
}

func (req *ListNetworkReq) Validate() apperrors.ValidationErrors {
	var validators []vld.Validator
	validators = append(validators, req.ListSettingReq.Validate()...)
	return apperrors.NewValidationErrors(vld.Validate(validators...))
}

type ListNetworkResp struct {
	Meta *basedto.ListMeta `json:"meta"`
	Data []*NetworkResp    `json:"data"`
}

func TransformNetworks(
	settings []*entity.Setting,
	refObjects *entity.RefObjects,
	refClusterObjects *entity.RefClusterObjects,
) ([]*NetworkResp, error) {
	resp := make([]*NetworkResp, 0, len(settings))
	for _, setting := range settings {
		item, err := TransformNetwork(setting, refObjects, refClusterObjects)
		if err != nil {
			return nil, apperrors.New(err)
		}
		resp = append(resp, item)
	}
	return resp, nil
}
