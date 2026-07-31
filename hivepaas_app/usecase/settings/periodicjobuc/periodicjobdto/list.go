package periodicjobdto

import (
	vld "github.com/tiendc/go-validator"

	"github.com/hivepaas/hivepaas/hivepaas_app/apperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/basedto"
	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/settings"
)

type ListPeriodicJobReq struct {
	settings.ListSettingReq
}

func NewListPeriodicJobReq() *ListPeriodicJobReq {
	return &ListPeriodicJobReq{
		ListSettingReq: settings.ListSettingReq{
			Paging: basedto.Paging{
				// Default paging if unset by client
				Sort: basedto.Orders{{Direction: basedto.DirectionAsc, ColumnName: "name"}},
			},
		},
	}
}

func (req *ListPeriodicJobReq) Validate() apperrors.ValidationErrors {
	var validators []vld.Validator
	validators = append(validators, req.ListSettingReq.Validate()...)
	return apperrors.NewValidationErrors(vld.Validate(validators...))
}

type ListPeriodicJobResp struct {
	Meta *basedto.ListMeta  `json:"meta"`
	Data []*PeriodicJobResp `json:"data"`
}

func TransformPeriodicJobs(
	settings []*entity.Setting,
	refObjects *entity.RefObjects,
) ([]*PeriodicJobResp, error) {
	resp := make([]*PeriodicJobResp, 0, len(settings))
	for _, setting := range settings {
		item, err := TransformPeriodicJob(setting, refObjects)
		if err != nil {
			return nil, apperrors.Wrap(err)
		}
		resp = append(resp, item)
	}
	return resp, nil
}
