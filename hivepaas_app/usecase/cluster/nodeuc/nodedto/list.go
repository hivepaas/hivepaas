package nodedto

import (
	vld "github.com/tiendc/go-validator"

	"github.com/hivepaas/hivepaas/hivepaas_app/apperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/basedto"
	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/settings"
)

type ListNodeReq struct {
	settings.ListSettingReq
}

func NewListNodeReq() *ListNodeReq {
	return &ListNodeReq{
		ListSettingReq: settings.ListSettingReq{
			Paging: basedto.Paging{
				// Default paging if unset by client
				Sort: basedto.Orders{{Direction: basedto.DirectionAsc, ColumnName: "name"}},
			},
		},
	}
}

func (req *ListNodeReq) Validate() apperrors.ValidationErrors {
	var validators []vld.Validator
	validators = append(validators, req.ListSettingReq.Validate()...)
	return apperrors.NewValidationErrors(vld.Validate(validators...))
}

type ListNodeResp struct {
	Meta *basedto.ListMeta `json:"meta"`
	Data []*NodeResp       `json:"data"`
}

func TransformNodes(
	settings []*entity.Setting,
	refObjects *entity.RefObjects,
	refClusterObjects *entity.RefClusterObjects,
	detailed bool,
) ([]*NodeResp, error) {
	resp := make([]*NodeResp, 0, len(settings))
	for _, setting := range settings {
		item, err := TransformNode(setting, refObjects, refClusterObjects, detailed)
		if err != nil {
			return nil, apperrors.New(err)
		}
		resp = append(resp, item)
	}
	return resp, nil
}
