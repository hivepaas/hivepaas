package volumedto

import (
	"time"

	"github.com/docker/docker/api/types/volume"
	vld "github.com/tiendc/go-validator"

	"github.com/localpaas/localpaas/localpaas_app/apperrors"
	"github.com/localpaas/localpaas/localpaas_app/base"
	"github.com/localpaas/localpaas/localpaas_app/basedto"
	"github.com/localpaas/localpaas/services/docker"
)

type GetVolumeReq struct {
	VolumeID  string `json:"-"`
	ProjectID string `json:"-"`
}

func NewGetVolumeReq() *GetVolumeReq {
	return &GetVolumeReq{}
}

func (req *GetVolumeReq) Validate() apperrors.ValidationErrors {
	var validators []vld.Validator
	// NOTE: volume id is docker id, it's not ULID
	validators = append(validators, basedto.ValidateStr(&req.VolumeID, true, 1, volumeIDMaxLen, "volumeID")...)
	validators = append(validators, basedto.ValidateID(&req.ProjectID, false, "projectID")...)
	return apperrors.NewValidationErrors(vld.Validate(validators...))
}

type GetVolumeResp struct {
	Meta *basedto.Meta `json:"meta"`
	Data *VolumeResp   `json:"data"`
}

type VolumeResp struct {
	ID                string                 `json:"id"`
	AvailInProjects   bool                   `json:"availableInProjects"`
	Labels            map[string]string      `json:"labels"`
	Driver            string                 `json:"driver"`
	Mountpoint        string                 `json:"mountpoint"`
	Options           map[string]string      `json:"options"`
	Scope             base.VolumeScope       `json:"scope"`
	Status            map[string]any         `json:"status"`
	RefCount          int64                  `json:"refCount"`
	Size              int64                  `json:"size"`
	CreatedAt         time.Time              `json:"createdAt"`
	ClusterVolumeSpec *ClusterVolumeSpecResp `json:"clusterVolumeSpec"`
	UpdateVer         int                    `json:"updateVer"`
}

type ClusterVolumeSpecResp struct {
	// TODO: add fields
}

func TransformVolume(vol *volume.Volume, _ bool) *VolumeResp {
	resp := &VolumeResp{
		ID:              vol.Name,
		AvailInProjects: vol.Labels[docker.StackLabelNamespace] == "",
		Driver:          vol.Driver,
		Mountpoint:      vol.Mountpoint,
		Options:         vol.Options,
		Scope:           base.VolumeScope(vol.Scope),
		Status:          vol.Status,
		Labels:          vol.Labels,
		CreatedAt:       transformVolumeCreatedAt(vol.CreatedAt),
	}
	if vol.ClusterVolume != nil {
		resp.ID = vol.ClusterVolume.ID
		resp.UpdateVer = int(vol.ClusterVolume.Version.Index) //nolint:gosec
	}
	if vol.UsageData != nil {
		resp.RefCount = vol.UsageData.RefCount
		resp.Size = vol.UsageData.Size
	}
	return resp
}

func transformVolumeCreatedAt(createdAt string) time.Time {
	t, err := time.Parse(time.RFC3339, createdAt)
	if err == nil {
		return t
	}
	return time.Time{}
}
