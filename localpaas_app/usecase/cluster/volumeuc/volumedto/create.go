package volumedto

import (
	vld "github.com/tiendc/go-validator"

	"github.com/localpaas/localpaas/localpaas_app/apperrors"
	"github.com/localpaas/localpaas/localpaas_app/basedto"
	"github.com/localpaas/localpaas/localpaas_app/pkg/unit"
	"github.com/localpaas/localpaas/services/docker"
)

const (
	volumeIDMaxLen   = 100
	volumeNameMaxLen = 100
)

type CreateVolumeReq struct {
	ProjectID       string              `json:"projectId"`
	AvailInProjects bool                `json:"availableInProjects"`
	Name            string              `json:"name"`
	Driver          docker.VolumeDriver `json:"driver"`

	// For `local` driver only
	NfsOptions   *VolumeNfsOptionsReq   `json:"nfsOptions"`
	TmpfsOptions *VolumeTmpfsOptionsReq `json:"tmpfsOptions"`
	BtrfsOptions *VolumeBtrfsOptionsReq `json:"btrfsOptions"`

	Options map[string]string `json:"options"`
	Labels  map[string]string `json:"labels"`
}

type VolumeNfsOptionsReq struct {
	Addr     string `json:"addr"`
	Device   string `json:"device"`
	Readonly bool   `json:"readonly"`
	Version  string `json:"version"`
}

func (req *VolumeNfsOptionsReq) validate(field string) (res []vld.Validator) {
	if req == nil {
		return res
	}
	if field != "" {
		field += "."
	}
	res = append(res, basedto.ValidateStr(&req.Addr, true, 1, volumeNameMaxLen, field+"addr")...)
	res = append(res, basedto.ValidateStr(&req.Device, true, 1, volumeNameMaxLen, field+"device")...)
	return res
}

type VolumeTmpfsOptionsReq struct {
	Size   unit.DataSize `json:"size"`
	UID    int           `json:"uid"`
	Device string        `json:"device"`
}

func (req *VolumeTmpfsOptionsReq) validate(_ string) []vld.Validator {
	return nil
}

type VolumeBtrfsOptionsReq struct {
	Device string `json:"device"`
}

func (req *VolumeBtrfsOptionsReq) validate(_ string) []vld.Validator {
	return nil
}

func NewCreateVolumeReq() *CreateVolumeReq {
	return &CreateVolumeReq{}
}

func (req *CreateVolumeReq) Validate() apperrors.ValidationErrors {
	var validators []vld.Validator
	validators = append(validators, basedto.ValidateID(&req.ProjectID, false, "projectId")...)
	validators = append(validators, basedto.ValidateStr(&req.Name, true, 1, volumeNameMaxLen, "name")...)
	validators = append(validators, req.NfsOptions.validate("nfsOpts")...)
	validators = append(validators, req.TmpfsOptions.validate("tmpfsOpts")...)
	validators = append(validators, req.BtrfsOptions.validate("btrfsOpts")...)

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
