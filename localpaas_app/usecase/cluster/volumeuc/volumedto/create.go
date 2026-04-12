package volumedto

import (
	vld "github.com/tiendc/go-validator"

	"github.com/localpaas/localpaas/localpaas_app/apperrors"
	"github.com/localpaas/localpaas/localpaas_app/base"
	"github.com/localpaas/localpaas/localpaas_app/basedto"
	"github.com/localpaas/localpaas/services/docker"
)

const (
	volumeIDMaxLen   = 100
	volumeNameMaxLen = 100
)

type CreateVolumeReq struct {
	ProjectID       string            `json:"projectId"`
	AvailInProjects bool              `json:"availableInProjects"`
	Name            string            `json:"name"`
	Driver          base.VolumeDriver `json:"driver"`
	Type            base.VolumeType   `json:"type"`
	Source          string            `json:"source"`
	NfsOpts         VolumeNfsOptsReq  `json:"nfsOpts"`
	ExtraDriverOpts map[string]string `json:"extraDriverOpts"`
	Labels          map[string]string `json:"labels"`
}

type VolumeNfsOptsReq struct {
	Addr     string `json:"addr"`
	Device   string `json:"device"`
	Readonly bool   `json:"readonly"`
	Version  string `json:"version"`
}

func (req *VolumeNfsOptsReq) validate(field string) (res []vld.Validator) {
	if field != "" {
		field += "."
	}
	res = append(res, basedto.ValidateStr(&req.Addr, true, 1, volumeNameMaxLen, field+"addr")...)
	res = append(res, basedto.ValidateStr(&req.Device, true, 1, volumeNameMaxLen, field+"device")...)
	return res
}

func NewCreateVolumeReq() *CreateVolumeReq {
	return &CreateVolumeReq{}
}

func (req *CreateVolumeReq) Validate() apperrors.ValidationErrors {
	var validators []vld.Validator
	validators = append(validators, basedto.ValidateID(&req.ProjectID, false, "projectId")...)
	validators = append(validators, basedto.ValidateStr(&req.Name, true, 1, volumeNameMaxLen, "name")...)
	validators = append(validators, basedto.ValidateStrIn(&req.Driver, true, base.AllVolumeDrivers, "driver")...)
	validators = append(validators, basedto.ValidateStrIn(&req.Type, true, base.AllVolumeTypes, "type")...)
	if req.Type == base.VolumeTypeNfs {
		validators = append(validators, req.NfsOpts.validate("nfsOpts")...)
	}

	// Validate volume labels
	unallowedLabels := docker.ValidateUserLabels(req.Labels, true)
	if len(unallowedLabels) > 0 {
		validators = append(validators, vld.Must(false).OnError(
			vld.SetField("labels", nil),
			vld.SetCustomKey("ERR_VLD_DOCKER_LABEL_UNALLOWED"),
			vld.SetParam("Label", unallowedLabels[0]),
		))
	}

	return apperrors.NewValidationErrors(vld.Validate(validators...))
}

type CreateVolumeResp struct {
	Meta *basedto.Meta         `json:"meta"`
	Data *basedto.ObjectIDResp `json:"data"`
}
