package emaildto

import (
	vld "github.com/tiendc/go-validator"

	"github.com/hivepaas/hivepaas/hivepaas_app/apperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/basedto"
	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/settings"
)

type ListEmailReq struct {
	settings.ListSettingReq
}

func NewListEmailReq() *ListEmailReq {
	return &ListEmailReq{
		ListSettingReq: settings.ListSettingReq{
			Paging: basedto.Paging{
				// Default paging if unset by client
				Sort: basedto.Orders{{Direction: basedto.DirectionAsc, ColumnName: "name"}},
			},
		},
	}
}

func (req *ListEmailReq) Validate() apperrors.ValidationErrors {
	var validators []vld.Validator
	validators = append(validators, req.ListSettingReq.Validate()...)
	return apperrors.NewValidationErrors(vld.Validate(validators...))
}

type ListEmailResp struct {
	Meta *basedto.ListMeta `json:"meta"`
	Data []*EmailResp      `json:"data"`
}

func TransformEmails(
	settings []*entity.Setting,
	refObjects *entity.RefObjects,
) (resp []*EmailResp, err error) {
	resp = make([]*EmailResp, 0, len(settings))
	for _, setting := range settings {
		item, err := TransformEmail(setting, refObjects)
		if err != nil {
			return nil, apperrors.Wrap(err)
		}
		resp = append(resp, item)
	}
	return resp, nil
}
