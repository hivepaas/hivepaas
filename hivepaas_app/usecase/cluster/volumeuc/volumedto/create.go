package volumedto

import (
	"math"

	"github.com/moby/moby/api/types/mount"
	vld "github.com/tiendc/go-validator"

	"github.com/hivepaas/hivepaas/hivepaas_app/basedto"
	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/fileutil"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/unit"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/settings"
	"github.com/hivepaas/hivepaas/services/docker"
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

	// Which node the volume's data is on, by id or by label.
	//
	// Empty is a claim, not an omission: it says the volume is reachable from
	// every node - a swarm cluster volume, or a bind whose directory comes from
	// shared storage mounted at the same path everywhere. HivePaaS cannot tell
	// that from a directory on one node's disk, so the caller has to say.
	//
	// What it decides: where a directory is created for a bind volume, and where
	// a backup repository kept on this volume runs. Leaving it empty when the
	// data is in fact on one node gives no directory and a backup repo that
	// refuses the volume.
	//
	// NodeID takes "current" to mean the node HivePaaS itself is running on.
	NodeID    string `json:"nodeId"`
	NodeLabel string `json:"nodeLabel"`

	// For `local` driver only
	BindOptions  *VolumeBindOptionsReq  `json:"bindOptions"`
	NfsOptions   *VolumeNfsOptionsReq   `json:"nfsOptions"`
	TmpfsOptions *VolumeTmpfsOptionsReq `json:"tmpfsOptions"`

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

	// One or the other, never both. They are two ways of naming the same thing -
	// where the data is - and the node exec service takes the id when both are
	// there, so a request carrying both would have its label quietly dropped
	// rather than answered. Neither one is still allowed: that is the claim that
	// the volume is reachable from every node.
	res = append(res, basedto.ValidateMutualExclusiveFields(
		req.NodeID == "" || req.NodeLabel == "",
		field+"nodeId", field+"nodeLabel")...)

	res = append(res, req.BindOptions.validate(field+"bindOptions")...)
	res = append(res, req.NfsOptions.validate(field+"nfsOptions")...)
	res = append(res, req.TmpfsOptions.validate(field+"tmpfsOptions")...)
	return res
}

func (req *VolumeBaseReq) ToEntity() *entity.ClusterVolume {
	return &entity.ClusterVolume{
		NodeID:    req.NodeID,
		NodeLabel: req.NodeLabel,
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

func NewCreateVolumeReq() *CreateVolumeReq {
	return &CreateVolumeReq{}
}

// Validate implements interface basedto.ReqValidator
func (req *CreateVolumeReq) Validate() hperrors.ValidationErrors {
	validators := make([]vld.Validator, 0, 10) //nolint:mnd
	validators = append(validators, req.CreateSettingReq.Validate()...)
	validators = append(validators, req.validate("")...)
	return hperrors.NewValidationErrors(vld.Validate(validators...))
}

type CreateVolumeResp struct {
	Meta *basedto.Meta         `json:"meta"`
	Data *basedto.ObjectIDResp `json:"data"`
}
