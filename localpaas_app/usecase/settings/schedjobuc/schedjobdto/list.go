package schedjobdto

import (
	vld "github.com/tiendc/go-validator"

	"github.com/localpaas/localpaas/localpaas_app/apperrors"
	"github.com/localpaas/localpaas/localpaas_app/basedto"
	"github.com/localpaas/localpaas/localpaas_app/entity"
	"github.com/localpaas/localpaas/localpaas_app/usecase/settings"
)

type ListSchedJobReq struct {
	settings.ListSettingReq
}

func NewListSchedJobReq() *ListSchedJobReq {
	return &ListSchedJobReq{
		ListSettingReq: settings.ListSettingReq{
			Paging: basedto.Paging{
				// Default paging if unset by client
				Sort: basedto.Orders{{Direction: basedto.DirectionAsc, ColumnName: "name"}},
			},
		},
	}
}

func (req *ListSchedJobReq) Validate() apperrors.ValidationErrors {
	var validators []vld.Validator
	validators = append(validators, req.ListSettingReq.Validate()...)
	return apperrors.NewValidationErrors(vld.Validate(validators...))
}

type ListSchedJobResp struct {
	Meta *basedto.ListMeta `json:"meta"`
	Data []*SchedJobResp   `json:"data"`
}

func TransformSchedJobs(
	settings []*entity.Setting,
	refObjects *entity.RefObjects,
) ([]*SchedJobResp, error) {
	resp := make([]*SchedJobResp, 0, len(settings))
	for _, setting := range settings {
		item, err := TransformSchedJob(setting, refObjects)
		if err != nil {
			return nil, apperrors.Wrap(err)
		}
		resp = append(resp, item)
	}
	return resp, nil
}
