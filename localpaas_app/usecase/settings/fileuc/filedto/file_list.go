package filedto

import (
	vld "github.com/tiendc/go-validator"

	"github.com/localpaas/localpaas/localpaas_app/apperrors"
	"github.com/localpaas/localpaas/localpaas_app/basedto"
	"github.com/localpaas/localpaas/localpaas_app/entity"
	"github.com/localpaas/localpaas/localpaas_app/usecase/settings"
)

type ListFileReq struct {
	settings.ListSettingReq
}

func NewListFileReq() *ListFileReq {
	return &ListFileReq{
		ListSettingReq: settings.ListSettingReq{
			Paging: basedto.Paging{
				// Default paging if unset by client
				Sort: basedto.Orders{{Direction: basedto.DirectionAsc, ColumnName: "name"}},
			},
		},
	}
}

func (req *ListFileReq) Validate() apperrors.ValidationErrors {
	var validators []vld.Validator
	validators = append(validators, req.ListSettingReq.Validate()...)
	return apperrors.NewValidationErrors(vld.Validate(validators...))
}

type ListFileResp struct {
	Meta *basedto.ListMeta `json:"meta"`
	Data []*FileResp       `json:"data"`
}

func TransformFiles(
	settings []*entity.Setting,
	refObjects *entity.RefObjects,
) (resp []*FileResp, err error) {
	resp = make([]*FileResp, 0, len(settings))
	for _, setting := range settings {
		item, err := TransformFile(setting, refObjects)
		if err != nil {
			return nil, apperrors.Wrap(err)
		}
		resp = append(resp, item)
	}
	return resp, nil
}
