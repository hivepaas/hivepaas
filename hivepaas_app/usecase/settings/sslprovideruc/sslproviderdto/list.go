package sslproviderdto

import (
	vld "github.com/tiendc/go-validator"

	"github.com/hivepaas/hivepaas/hivepaas_app/apperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/basedto"
	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/settings"
)

type ListSSLProviderReq struct {
	settings.ListSettingReq
}

func NewListSSLProviderReq() *ListSSLProviderReq {
	return &ListSSLProviderReq{
		ListSettingReq: settings.ListSettingReq{
			Paging: basedto.Paging{
				// Default paging if unset by client
				Sort: basedto.Orders{{Direction: basedto.DirectionAsc, ColumnName: "name"}},
			},
		},
	}
}

func (req *ListSSLProviderReq) Validate() apperrors.ValidationErrors {
	var validators []vld.Validator
	validators = append(validators, req.ListSettingReq.Validate()...)
	return apperrors.NewValidationErrors(vld.Validate(validators...))
}

type ListSSLProviderResp struct {
	Meta *basedto.ListMeta  `json:"meta"`
	Data []*SSLProviderResp `json:"data"`
}

func TransformSSLProviders(
	settings []*entity.Setting,
	refObjects *entity.RefObjects,
) (resp []*SSLProviderResp, err error) {
	resp = make([]*SSLProviderResp, 0, len(settings))
	for _, setting := range settings {
		item, err := TransformSSLProvider(setting, refObjects)
		if err != nil {
			return nil, apperrors.Wrap(err)
		}
		resp = append(resp, item)
	}
	return resp, nil
}
