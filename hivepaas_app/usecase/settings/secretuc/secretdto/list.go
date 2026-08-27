package secretdto

import (
	vld "github.com/tiendc/go-validator"

	"github.com/hivepaas/hivepaas/hivepaas_app/basedto"
	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/settings"
)

type ListSecretReq struct {
	settings.ListSettingReq
}

func NewListSecretReq() *ListSecretReq {
	return &ListSecretReq{
		ListSettingReq: settings.ListSettingReq{
			Paging: basedto.Paging{
				// Default paging if unset by client
				Sort: basedto.Orders{{Direction: basedto.DirectionAsc, ColumnName: "name"}},
			},
		},
	}
}

func (req *ListSecretReq) Validate() hperrors.ValidationErrors {
	validators := make([]vld.Validator, 0, 10) //nolint:mnd
	validators = append(validators, req.ListSettingReq.Validate()...)
	return hperrors.NewValidationErrors(vld.Validate(validators...))
}

type ListSecretResp struct {
	Meta *basedto.ListMeta `json:"meta"`
	Data []*SecretResp     `json:"data"`
}

func TransformSecrets(
	settings []*entity.Setting,
	refObjects *entity.RefObjects,
) (resp []*SecretResp, err error) {
	resp = make([]*SecretResp, 0, len(settings))
	for _, setting := range settings {
		item, err := TransformSecret(setting, refObjects)
		if err != nil {
			return nil, hperrors.Wrap(err)
		}
		resp = append(resp, item)
	}
	return resp, nil
}
