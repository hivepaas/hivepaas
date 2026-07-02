package oauthdto

import (
	vld "github.com/tiendc/go-validator"

	"github.com/hivepaas/hivepaas/hivepaas_app/apperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/basedto"
	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/settings"
)

type ListOAuthReq struct {
	settings.ListSettingReq
	Kind []base.OAuthKind `json:"-" mapstructure:"kind"`
}

func NewListOAuthReq() *ListOAuthReq {
	return &ListOAuthReq{
		ListSettingReq: settings.ListSettingReq{
			Paging: basedto.Paging{
				// Default paging if unset by client
				Sort: basedto.Orders{{Direction: basedto.DirectionAsc, ColumnName: "name"}},
			},
		},
	}
}

func (req *ListOAuthReq) Validate() apperrors.ValidationErrors {
	var validators []vld.Validator
	validators = append(validators, req.ListSettingReq.Validate()...)
	validators = append(validators, basedto.ValidateSlice(req.Kind, true, 0, base.AllOAuthKinds, "kind")...)
	return apperrors.NewValidationErrors(vld.Validate(validators...))
}

type ListOAuthResp struct {
	Meta *basedto.ListMeta `json:"meta"`
	Data []*OAuthResp      `json:"data"`
}

func TransformOAuths(
	settings []*entity.Setting,
	input *OAuthTransformInput,
) ([]*OAuthResp, error) {
	resp := make([]*OAuthResp, 0, len(settings))
	for _, setting := range settings {
		item, err := TransformOAuth(setting, input)
		if err != nil {
			return nil, apperrors.New(err)
		}
		resp = append(resp, item)
	}
	return resp, nil
}
