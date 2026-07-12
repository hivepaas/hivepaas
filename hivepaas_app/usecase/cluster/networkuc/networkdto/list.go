package networkdto

import (
	vld "github.com/tiendc/go-validator"

	"github.com/hivepaas/hivepaas/hivepaas_app/apperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/basedto"
	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/settings"
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
			return nil, apperrors.Wrap(err)
		}
		resp = append(resp, item)
	}
	return resp, nil
}
