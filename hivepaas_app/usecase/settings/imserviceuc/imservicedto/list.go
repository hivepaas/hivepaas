package imservicedto

import (
	vld "github.com/tiendc/go-validator"

	"github.com/hivepaas/hivepaas/hivepaas_app/basedto"
	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/settings"
)

type ListIMServiceReq struct {
	settings.ListSettingReq
}

func NewListIMServiceReq() *ListIMServiceReq {
	return &ListIMServiceReq{
		ListSettingReq: settings.ListSettingReq{
			Paging: basedto.Paging{
				// Default paging if unset by client
				Sort: basedto.Orders{{Direction: basedto.DirectionAsc, ColumnName: "name"}},
			},
		},
	}
}

func (req *ListIMServiceReq) Validate() hperrors.ValidationErrors {
	validators := make([]vld.Validator, 0, 10) //nolint:mnd
	validators = append(validators, req.ListSettingReq.Validate()...)
	return hperrors.NewValidationErrors(vld.Validate(validators...))
}

type ListIMServiceResp struct {
	Meta *basedto.ListMeta `json:"meta"`
	Data []*IMServiceResp  `json:"data"`
}

func TransformIMServices(
	settings []*entity.Setting,
	refObjects *entity.RefObjects,
) (resp []*IMServiceResp, err error) {
	resp = make([]*IMServiceResp, 0, len(settings))
	for _, setting := range settings {
		item, err := TransformIMService(setting, refObjects)
		if err != nil {
			return nil, hperrors.Wrap(err)
		}
		resp = append(resp, item)
	}
	return resp, nil
}
