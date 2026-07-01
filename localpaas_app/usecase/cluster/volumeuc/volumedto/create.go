package volumedto

import (
	"math"

	"github.com/moby/moby/api/types/mount"
	vld "github.com/tiendc/go-validator"

	"github.com/localpaas/localpaas/localpaas_app/apperrors"
	"github.com/localpaas/localpaas/localpaas_app/basedto"
	"github.com/localpaas/localpaas/localpaas_app/entity"
	"github.com/localpaas/localpaas/localpaas_app/pkg/fileutil"
	"github.com/localpaas/localpaas/localpaas_app/pkg/unit"
	"github.com/localpaas/localpaas/localpaas_app/usecase/settings"
	"github.com/localpaas/localpaas/services/docker"
)

const (
	volumeNameMaxLen   = 100
	volumeDeviceMaxLen = 200
)

type CreateVolumeReq struct {
	settings.CreateSettingReq
	*VolumeBaseReq
}

type VolumeBaseReq struct {
	Name   string              `json:"name"`
	Driver docker.VolumeDriver `json:"driver"`

	// For `local` driver only
	BindOptions  *VolumeBindOptionsReq  `json:"bindOptions"`
	NfsOptions   *VolumeNfsOptionsReq   `json:"nfsOptions"`
	TmpfsOptions *VolumeTmpfsOptionsReq `json:"tmpfsOptions"`
	BtrfsOptions *VolumeBtrfsOptionsReq `json:"btrfsOptions"`

	Options map[string]string `json:"options"`
	Labels  map[string]string `json:"labels"`
}

func (req *VolumeBaseReq) validate(field string) (res []vld.Validator) {
	if req == nil {
		return res
	}
	if field != "" {
		field += "."
	}
	res = append(res, basedto.ValidateStr(&req.Name, true, 1, volumeNameMaxLen, field+"name")...)
	res = append(res, req.BindOptions.validate(field+"bindOptions")...)
	res = append(res, req.NfsOptions.validate(field+"nfsOptions")...)
	res = append(res, req.TmpfsOptions.validate(field+"tmpfsOptions")...)
	res = append(res, req.BtrfsOptions.validate(field+"btrfsOptions")...)
	return res
}

func (req *VolumeBaseReq) ToEntity() *entity.ClusterVolume {
	return &entity.ClusterVolume{
		Name:   req.Name,
		Driver: req.Driver,
	}
}

type VolumeBindOptionsReq struct {
	Directory    string            `json:"directory"`
	Propagation  mount.Propagation `json:"propagation"`
	Readonly     bool              `json:"readonly"`
	ExtraOptions string            `json:"extraOptions"`
}

func (req *VolumeBindOptionsReq) validate(field string) (res []vld.Validator) {
	if req == nil {
		return res
	}
	if field != "" {
		field += "."
	}
	res = append(res, basedto.ValidateStr(&req.Directory, false, 1, volumeDeviceMaxLen, field+"directory")...)
	res = append(res, basedto.ValidateStrIn(&req.Propagation, false, mount.Propagations, field+"propagation")...)
	return res
}

type VolumeNfsOptionsReq struct {
	Addr         string `json:"addr"`
	Device       string `json:"device"`
	Readonly     bool   `json:"readonly"`
	Version      string `json:"version"`
	ExtraOptions string `json:"extraOptions"`
}

func (req *VolumeNfsOptionsReq) validate(field string) (res []vld.Validator) {
	if req == nil {
		return res
	}
	if field != "" {
		field += "."
	}
	res = append(res, basedto.ValidateStr(&req.Addr, true, 1, volumeNameMaxLen, field+"addr")...)
	res = append(res, basedto.ValidateStr(&req.Device, true, 1, volumeDeviceMaxLen, field+"device")...)
	return res
}

type VolumeTmpfsOptionsReq struct {
	Size         unit.DataSize     `json:"size"`
	Mode         fileutil.FileMode `json:"mode"`
	UID          int               `json:"uid"`
	GID          int               `json:"gid"`
	Device       string            `json:"device"`
	ExtraOptions string            `json:"extraOptions"`
}

func (req *VolumeTmpfsOptionsReq) validate(field string) (res []vld.Validator) {
	if req == nil {
		return res
	}
	if field != "" {
		field += "."
	}
	res = append(res, basedto.ValidateNumber(&req.Size, true, unit.MB, math.MaxInt64, field+"size")...)
	res = append(res, basedto.ValidateStr(&req.Device, false, 1, volumeDeviceMaxLen, field+"device")...)
	return res
}

type VolumeBtrfsOptionsReq struct {
	Device       string `json:"device"`
	ExtraOptions string `json:"extraOptions"`
}

func (req *VolumeBtrfsOptionsReq) validate(field string) (res []vld.Validator) {
	if req == nil {
		return res
	}
	if field != "" {
		field += "."
	}
	res = append(res, basedto.ValidateStr(&req.Device, true, 1, volumeDeviceMaxLen, field+"device")...)
	return res
}

func NewCreateVolumeReq() *CreateVolumeReq {
	return &CreateVolumeReq{}
}

// Validate implements interface basedto.ReqValidator
func (req *CreateVolumeReq) Validate() apperrors.ValidationErrors {
	var validators []vld.Validator
	validators = append(validators, req.validate("")...)
	return apperrors.NewValidationErrors(vld.Validate(validators...))
}

type CreateVolumeResp struct {
	Meta *basedto.Meta         `json:"meta"`
	Data *basedto.ObjectIDResp `json:"data"`
}
