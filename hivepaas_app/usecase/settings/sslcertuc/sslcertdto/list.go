package sslcertdto

import (
	vld "github.com/tiendc/go-validator"

	"github.com/hivepaas/hivepaas/hivepaas_app/apperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/basedto"
	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/settings"
)

type ListSSLCertReq struct {
	settings.ListSettingReq
	Domain string `json:"-" mapstructure:"domain"`
}

func NewListSSLCertReq() *ListSSLCertReq {
	return &ListSSLCertReq{
		ListSettingReq: settings.ListSettingReq{
			Paging: basedto.Paging{
				// Default paging if unset by client
				Sort: basedto.Orders{{Direction: basedto.DirectionAsc, ColumnName: "name"}},
			},
		},
	}
}

func (req *ListSSLCertReq) Validate() apperrors.ValidationErrors {
	var validators []vld.Validator
	validators = append(validators, req.ListSettingReq.Validate()...)
	return apperrors.NewValidationErrors(vld.Validate(validators...))
}

type ListSSLCertResp struct {
	Meta *basedto.ListMeta `json:"meta"`
	Data []*SSLCertResp    `json:"data"`
}

func TransformSSLCerts(
	settings []*entity.Setting,
	refObjects *entity.RefObjects,
) (resp []*SSLCertResp, err error) {
	resp = make([]*SSLCertResp, 0, len(settings))
	for _, setting := range settings {
		item, err := TransformSSLCert(setting, refObjects)
		if err != nil {
			return nil, apperrors.New(err)
		}
		resp = append(resp, item)
	}
	return resp, nil
}
