package appdto

import (
	"github.com/docker/docker/api/types/mount"
	"github.com/docker/docker/api/types/swarm"
	vld "github.com/tiendc/go-validator"

	"github.com/localpaas/localpaas/localpaas_app/apperrors"
	"github.com/localpaas/localpaas/localpaas_app/basedto"
)

type GetAppStorageSettingsReq struct {
	ProjectID string `json:"-"`
	AppID     string `json:"-"`
}

func NewGetAppStorageSettingsReq() *GetAppStorageSettingsReq {
	return &GetAppStorageSettingsReq{}
}

func (req *GetAppStorageSettingsReq) Validate() apperrors.ValidationErrors {
	var validators []vld.Validator
	validators = append(validators, basedto.ValidateID(&req.ProjectID, true, "projectID")...)
	validators = append(validators, basedto.ValidateID(&req.AppID, true, "appID")...)
	return apperrors.NewValidationErrors(vld.Validate(validators...))
}

type GetAppStorageSettingsResp struct {
	Meta *basedto.Meta        `json:"meta"`
	Data *StorageSettingsResp `json:"data"`
}

type StorageSettingsResp struct {
	Mounts []*Mount `json:"mounts"`

	UpdateVer int `json:"updateVer"`
}

type Mount struct {
	Type        mount.Type        `json:"type"`
	Source      string            `json:"source,omitempty"`
	Target      string            `json:"target,omitempty"`
	ReadOnly    bool              `json:"readOnly,omitempty"`
	Consistency mount.Consistency `json:"consistency,omitempty"`
}

func TransformStorageSettings(
	service *swarm.Service,
) (resp *StorageSettingsResp, err error) {
	spec := &service.Spec
	resp = &StorageSettingsResp{
		UpdateVer: int(service.Version.Index), //nolint:gosec
	}

	resp.Mounts = TransformVolumeMounts(spec.TaskTemplate.ContainerSpec.Mounts)

	return resp, nil
}

func TransformVolumeMounts(mounts []mount.Mount) []*Mount {
	resp := make([]*Mount, 0, len(mounts))
	for _, mnt := range mounts {
		resp = append(resp, &Mount{
			Type:        mnt.Type,
			Source:      mnt.Source,
			Target:      mnt.Target,
			ReadOnly:    mnt.ReadOnly,
			Consistency: mnt.Consistency,
		})
	}
	return resp
}
