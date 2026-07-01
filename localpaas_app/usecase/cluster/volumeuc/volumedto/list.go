package volumedto

import (
	vld "github.com/tiendc/go-validator"

	"github.com/localpaas/localpaas/localpaas_app/apperrors"
	"github.com/localpaas/localpaas/localpaas_app/basedto"
	"github.com/localpaas/localpaas/localpaas_app/entity"
	"github.com/localpaas/localpaas/localpaas_app/usecase/settings"
)

type ListVolumeReq struct {
	settings.ListSettingReq
}

func NewListVolumeReq() *ListVolumeReq {
	return &ListVolumeReq{
		ListSettingReq: settings.ListSettingReq{
			Paging: basedto.Paging{
				// Default paging if unset by client
				Sort: basedto.Orders{{Direction: basedto.DirectionAsc, ColumnName: "name"}},
			},
		},
	}
}

func (req *ListVolumeReq) Validate() apperrors.ValidationErrors {
	var validators []vld.Validator
	validators = append(validators, req.ListSettingReq.Validate()...)
	return apperrors.NewValidationErrors(vld.Validate(validators...))
}

type ListVolumeResp struct {
	Meta *basedto.ListMeta `json:"meta"`
	Data []*VolumeResp     `json:"data"`
}

func TransformVolumes(
	settings []*entity.Setting,
	refObjects *entity.RefObjects,
	refClusterObjects *entity.RefClusterObjects,
) ([]*VolumeResp, error) {
	resp := make([]*VolumeResp, 0, len(settings))
	for _, setting := range settings {
		item, err := TransformVolume(setting, refObjects, refClusterObjects)
		if err != nil {
			return nil, apperrors.New(err)
		}
		resp = append(resp, item)
	}
	return resp, nil
}
