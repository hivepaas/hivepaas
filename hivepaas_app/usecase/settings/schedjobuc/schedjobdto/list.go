package schedjobdto

import (
	vld "github.com/tiendc/go-validator"

	"github.com/hivepaas/hivepaas/hivepaas_app/apperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/basedto"
	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/settings"
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
		item, err := TransformSchedJob(setting, refObjects, true)
		if err != nil {
			return nil, apperrors.New(err)
		}
		resp = append(resp, item)
	}
	return resp, nil
}
