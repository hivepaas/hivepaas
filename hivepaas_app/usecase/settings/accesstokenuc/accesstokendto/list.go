package accesstokendto

import (
	vld "github.com/tiendc/go-validator"

	"github.com/hivepaas/hivepaas/hivepaas_app/basedto"
	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/settings"
)

type ListAccessTokenReq struct {
	settings.ListSettingReq
}

func NewListAccessTokenReq() *ListAccessTokenReq {
	return &ListAccessTokenReq{
		ListSettingReq: settings.ListSettingReq{
			Paging: basedto.Paging{
				// Default paging if unset by client
				Sort: basedto.Orders{{Direction: basedto.DirectionAsc, ColumnName: "name"}},
			},
		},
	}
}

func (req *ListAccessTokenReq) Validate() hperrors.ValidationErrors {
	validators := make([]vld.Validator, 0, 10) //nolint:mnd
	validators = append(validators, req.ListSettingReq.Validate()...)
	return hperrors.NewValidationErrors(vld.Validate(validators...))
}

type ListAccessTokenResp struct {
	Meta *basedto.ListMeta  `json:"meta"`
	Data []*AccessTokenResp `json:"data"`
}

func TransformAccessTokens(
	settings []*entity.Setting,
	refObjects *entity.RefObjects,
) (resp []*AccessTokenResp, err error) {
	resp = make([]*AccessTokenResp, 0, len(settings))
	for _, setting := range settings {
		item, err := TransformAccessToken(setting, refObjects)
		if err != nil {
			return nil, hperrors.Wrap(err)
		}
		resp = append(resp, item)
	}
	return resp, nil
}
