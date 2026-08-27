package commandpipedto

import (
	vld "github.com/tiendc/go-validator"

	"github.com/hivepaas/hivepaas/hivepaas_app/basedto"
	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/settings"
)

type ListCommandPipeReq struct {
	settings.ListSettingReq
}

func NewListCommandPipeReq() *ListCommandPipeReq {
	return &ListCommandPipeReq{
		ListSettingReq: settings.ListSettingReq{
			Paging: basedto.Paging{
				Sort: basedto.Orders{{Direction: basedto.DirectionAsc, ColumnName: "name"}},
			},
		},
	}
}

func (req *ListCommandPipeReq) Validate() hperrors.ValidationErrors {
	validators := make([]vld.Validator, 0, 10) //nolint:mnd
	validators = append(validators, req.ListSettingReq.Validate()...)
	return hperrors.NewValidationErrors(vld.Validate(validators...))
}

type ListCommandPipeResp struct {
	Meta *basedto.ListMeta  `json:"meta"`
	Data []*CommandPipeResp `json:"data"`
}

func TransformCommandPipes(
	settings []*entity.Setting,
	refObjects *entity.RefObjects,
) (resp []*CommandPipeResp, err error) {
	resp = make([]*CommandPipeResp, 0, len(settings))
	for _, setting := range settings {
		item, err := TransformCommandPipe(setting, refObjects)
		if err != nil {
			return nil, hperrors.Wrap(err)
		}
		resp = append(resp, item)
	}
	return resp, nil
}
