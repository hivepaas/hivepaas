package keyauthdto

import (
	vld "github.com/tiendc/go-validator"

	"github.com/hivepaas/hivepaas/hivepaas_app/basedto"
	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/settings"
)

type ListKeyAuthReq struct {
	settings.ListSettingReq
}

func NewListKeyAuthReq() *ListKeyAuthReq {
	return &ListKeyAuthReq{
		ListSettingReq: settings.ListSettingReq{
			Paging: basedto.Paging{
				// Default paging if unset by client
				Sort: basedto.Orders{{Direction: basedto.DirectionAsc, ColumnName: "name"}},
			},
		},
	}
}

func (req *ListKeyAuthReq) Validate() hperrors.ValidationErrors {
	validators := make([]vld.Validator, 0, 10) //nolint:mnd
	validators = append(validators, req.ListSettingReq.Validate()...)
	return hperrors.NewValidationErrors(vld.Validate(validators...))
}

type ListKeyAuthResp struct {
	Meta *basedto.ListMeta `json:"meta"`
	Data []*KeyAuthResp    `json:"data"`
}

func TransformKeyAuths(
	settings []*entity.Setting,
	refObjects *entity.RefObjects,
) (resp []*KeyAuthResp, err error) {
	resp = make([]*KeyAuthResp, 0, len(settings))
	for _, setting := range settings {
		item, err := TransformKeyAuth(setting, refObjects)
		if err != nil {
			return nil, hperrors.Wrap(err)
		}
		resp = append(resp, item)
	}
	return resp, nil
}
