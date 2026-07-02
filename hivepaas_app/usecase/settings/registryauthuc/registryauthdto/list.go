package registryauthdto

import (
	vld "github.com/tiendc/go-validator"

	"github.com/hivepaas/hivepaas/hivepaas_app/apperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/basedto"
	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/settings"
)

type ListRegistryAuthReq struct {
	settings.ListSettingReq
}

func NewListRegistryAuthReq() *ListRegistryAuthReq {
	return &ListRegistryAuthReq{
		ListSettingReq: settings.ListSettingReq{
			Paging: basedto.Paging{
				// Default paging if unset by client
				Sort: basedto.Orders{{Direction: basedto.DirectionAsc, ColumnName: "name"}},
			},
		},
	}
}

func (req *ListRegistryAuthReq) Validate() apperrors.ValidationErrors {
	var validators []vld.Validator
	validators = append(validators, req.ListSettingReq.Validate()...)
	return apperrors.NewValidationErrors(vld.Validate(validators...))
}

type ListRegistryAuthResp struct {
	Meta *basedto.ListMeta   `json:"meta"`
	Data []*RegistryAuthResp `json:"data"`
}

func TransformRegistryAuths(
	settings []*entity.Setting,
	refObjects *entity.RefObjects,
) (resp []*RegistryAuthResp, err error) {
	resp = make([]*RegistryAuthResp, 0, len(settings))
	for _, setting := range settings {
		item, err := TransformRegistryAuth(setting, refObjects)
		if err != nil {
			return nil, apperrors.New(err)
		}
		resp = append(resp, item)
	}
	return resp, nil
}
