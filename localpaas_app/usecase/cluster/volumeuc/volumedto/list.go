package volumedto

import (
	"github.com/docker/docker/api/types/volume"
	vld "github.com/tiendc/go-validator"
	"github.com/tiendc/gofn"

	"github.com/localpaas/localpaas/localpaas_app/apperrors"
	"github.com/localpaas/localpaas/localpaas_app/basedto"
	"github.com/localpaas/localpaas/services/docker"
)

type ListVolumeReq struct {
	ProjectID string            `json:"-"`
	Type      docker.VolumeType `json:"-" mapstructure:"type"`
	ListAll   bool              `json:"-" mapstructure:"listAll"`
	Search    string            `json:"-" mapstructure:"search"`

	Paging basedto.Paging `json:"-"`
}

func NewListVolumeReq() *ListVolumeReq {
	return &ListVolumeReq{
		Paging: basedto.Paging{
			// Default paging if unset by client
			Sort: basedto.Orders{{Direction: basedto.DirectionAsc, ColumnName: "name"}},
		},
	}
}

func (req *ListVolumeReq) Validate() apperrors.ValidationErrors {
	var validators []vld.Validator
	validators = append(validators, basedto.ValidateID(&req.ProjectID, false, "projectId")...)
	validators = append(validators, basedto.ValidateStrIn(&req.Type, false, docker.AllVolumeTypes, "type")...)
	return apperrors.NewValidationErrors(vld.Validate(validators...))
}

type ListVolumeResp struct {
	Meta *basedto.ListMeta `json:"meta"`
	Data []*VolumeResp     `json:"data"`
}

func TransformVolumes(volumes []*volume.Volume, detailed bool) []*VolumeResp {
	return gofn.MapSlice(volumes, func(vol *volume.Volume) *VolumeResp {
		return TransformVolume(vol, detailed)
	})
}
