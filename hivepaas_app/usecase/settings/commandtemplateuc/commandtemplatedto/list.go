package commandtemplatedto

import (
	vld "github.com/tiendc/go-validator"

	"github.com/hivepaas/hivepaas/hivepaas_app/apperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/basedto"
	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/settings"
)

type ListCommandTemplateReq struct {
	settings.ListSettingReq
}

func NewListCommandTemplateReq() *ListCommandTemplateReq {
	return &ListCommandTemplateReq{
		ListSettingReq: settings.ListSettingReq{
			Paging: basedto.Paging{
				Sort: basedto.Orders{{Direction: basedto.DirectionAsc, ColumnName: "name"}},
			},
		},
	}
}

func (req *ListCommandTemplateReq) Validate() apperrors.ValidationErrors {
	var validators []vld.Validator
	validators = append(validators, req.ListSettingReq.Validate()...)
	return apperrors.NewValidationErrors(vld.Validate(validators...))
}

type ListCommandTemplateResp struct {
	Meta *basedto.ListMeta      `json:"meta"`
	Data []*CommandTemplateResp `json:"data"`
}

func TransformCommandTemplates(
	settings []*entity.Setting,
	refObjects *entity.RefObjects,
) (resp []*CommandTemplateResp, err error) {
	resp = make([]*CommandTemplateResp, 0, len(settings))
	for _, setting := range settings {
		item, err := TransformCommandTemplate(setting, refObjects)
		if err != nil {
			return nil, apperrors.New(err)
		}
		resp = append(resp, item)
	}
	return resp, nil
}
