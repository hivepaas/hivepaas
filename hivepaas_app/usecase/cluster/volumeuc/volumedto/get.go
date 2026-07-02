package volumedto

import (
	"strconv"
	"strings"
	"time"

	"github.com/moby/moby/api/types/mount"
	"github.com/moby/moby/api/types/volume"
	vld "github.com/tiendc/go-validator"
	"github.com/tiendc/gofn"

	"github.com/hivepaas/hivepaas/hivepaas_app/apperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/basedto"
	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/copier"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/fileutil"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/unit"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/settings"
	"github.com/hivepaas/hivepaas/services/docker"
)

type GetVolumeReq struct {
	settings.GetSettingReq
}

func NewGetVolumeReq() *GetVolumeReq {
	return &GetVolumeReq{}
}

func (req *GetVolumeReq) Validate() apperrors.ValidationErrors {
	var validators []vld.Validator
	validators = append(validators, req.GetSettingReq.Validate()...)
	return apperrors.NewValidationErrors(vld.Validate(validators...))
}

type GetVolumeResp struct {
	Meta *basedto.Meta `json:"meta"`
	Data *VolumeResp   `json:"data"`
}

type VolumeResp struct {
	*settings.BaseSettingResp

	Driver            docker.VolumeDriver    `json:"driver"`
	Scope             docker.VolumeScope     `json:"scope"`
	Mountpoint        string                 `json:"mountpoint"`
	Options           map[string]string      `json:"options"`
	Labels            map[string]string      `json:"labels"`
	RefCount          int64                  `json:"refCount"`
	ClusterVolumeSpec *ClusterVolumeSpecResp `json:"clusterVolumeSpec"`

	// If driver is `local`
	BindOptions  *VolumeBindOptionsResp  `json:"bindOptions,omitempty"`
	NfsOptions   *VolumeNfsOptionsResp   `json:"nfsOptions,omitempty"`
	TmpfsOptions *VolumeTmpfsOptionsResp `json:"tmpfsOptions,omitempty"`
	BtrfsOptions *VolumeBtrfsOptionsResp `json:"btrfsOptions,omitempty"`
}

type VolumeBindOptionsResp struct {
	Directory    string            `json:"directory"`
	Propagation  mount.Propagation `json:"propagation"`
	Readonly     bool              `json:"readonly"`
	ExtraOptions string            `json:"extraOptions"`
}

type VolumeNfsOptionsResp struct {
	Addr         string `json:"addr"`
	Device       string `json:"device"`
	Readonly     bool   `json:"readonly"`
	Version      string `json:"version"`
	ExtraOptions string `json:"extraOptions"`
}

type VolumeTmpfsOptionsResp struct {
	Size         unit.DataSize     `json:"size"`
	Mode         fileutil.FileMode `json:"mode"`
	UID          int               `json:"uid"`
	GID          int               `json:"gid"`
	Device       string            `json:"device"`
	ExtraOptions string            `json:"extraOptions"`
}

type VolumeBtrfsOptionsResp struct {
	Device       string `json:"device"`
	ExtraOptions string `json:"extraOptions"`
}

type ClusterVolumeSpecResp struct {
	// TODO: add fields
}

func TransformVolume(
	setting *entity.Setting,
	_ *entity.RefObjects,
	refClusterObjects *entity.RefClusterObjects,
) (resp *VolumeResp, err error) {
	volEnt := setting.MustAsClusterVolume()
	if err = copier.Copy(&resp, volEnt); err != nil {
		return nil, apperrors.New(err)
	}

	resp.BaseSettingResp, err = settings.TransformSettingBase(setting)
	if err != nil {
		return nil, apperrors.New(err)
	}

	vol := refClusterObjects.RefVolumes[volEnt.VolumeID]

	resp.Driver = docker.VolumeDriver(vol.Driver)
	resp.Mountpoint = vol.Mountpoint
	resp.Options = vol.Options
	resp.Scope = docker.VolumeScope(vol.Scope)
	resp.Labels = vol.Labels
	resp.CreatedAt = transformVolumeCreatedAt(vol.CreatedAt)
	if vol.ClusterVolume != nil {
		resp.ID = vol.ClusterVolume.ID
		resp.UpdateVer = int(vol.ClusterVolume.Version.Index) //nolint:gosec
	}
	if vol.UsageData != nil {
		resp.RefCount = vol.UsageData.RefCount
		resp.Size = vol.UsageData.Size
	}

	if resp.Driver == docker.VolumeDriverLocal {
		switch vol.Options["type"] {
		case "none":
			resp.BindOptions = transformVolumeTypeBind(vol)
			resp.Options = nil
		case "nfs":
			resp.NfsOptions = transformVolumeTypeNfs(vol)
			resp.Options = nil
		case "tmpfs":
			resp.TmpfsOptions = transformVolumeTypeTmpfs(vol)
			resp.Options = nil
		case "btrfs":
			resp.BtrfsOptions = transformVolumeTypeBtrfs(vol)
			resp.Options = nil
		}
	}
	return resp, nil
}

func transformVolumeCreatedAt(createdAt string) time.Time {
	t, err := time.Parse(time.RFC3339, createdAt)
	if err == nil {
		return t
	}
	return time.Time{}
}

func transformVolumeTypeBind(vol *volume.Volume) *VolumeBindOptionsResp {
	opts := strings.Split(vol.Options["o"], ",")
	if !gofn.Contain(opts, "bind") {
		return nil
	}
	resp := &VolumeBindOptionsResp{
		Directory: vol.Options["device"],
	}
	for _, opt := range opts {
		if opt == "bind" {
			continue
		}
		if opt == "ro" || opt == "rw" {
			resp.Readonly = opt == "ro"
			continue
		}
		if gofn.Contain(mount.Propagations, mount.Propagation(opt)) {
			resp.Propagation = mount.Propagation(opt)
			continue
		}
		if resp.ExtraOptions == "" {
			resp.ExtraOptions = opt
		} else {
			resp.ExtraOptions += "," + opt
		}
	}
	return resp
}

func transformVolumeTypeNfs(vol *volume.Volume) *VolumeNfsOptionsResp {
	resp := &VolumeNfsOptionsResp{
		Device: vol.Options["device"],
	}
	opts := strings.Split(vol.Options["o"], ",")
	for _, opt := range opts {
		if opt == "ro" || opt == "rw" {
			resp.Readonly = opt == "ro"
			continue
		}
		if strings.HasPrefix(opt, "addr=") {
			resp.Addr = strings.TrimPrefix(opt, "addr=")
			continue
		}
		if strings.HasPrefix(opt, "nfsvers=") {
			resp.Version = strings.TrimPrefix(opt, "nfsvers=")
			continue
		}
		if resp.ExtraOptions == "" {
			resp.ExtraOptions = opt
		} else {
			resp.ExtraOptions += "," + opt
		}
	}
	return resp
}

func transformVolumeTypeTmpfs(vol *volume.Volume) *VolumeTmpfsOptionsResp {
	resp := &VolumeTmpfsOptionsResp{
		Device: vol.Options["device"],
	}
	opts := strings.Split(vol.Options["o"], ",")
	for _, opt := range opts {
		if strings.HasPrefix(opt, "size=") {
			val := opt[len("size="):]
			if !strings.HasSuffix(val, "b") {
				val += "b"
			}
			if sz, err := unit.ParseDataSizeString(val); err == nil {
				resp.Size = sz
			}
			continue
		}
		if strings.HasPrefix(opt, "mode=") {
			val := opt[len("mode="):]
			if md, err := fileutil.ParseFileMode(val); err == nil {
				resp.Mode = md
			}
			continue
		}
		if strings.HasPrefix(opt, "uid=") {
			val := opt[len("uid="):]
			if uidVal, err := strconv.Atoi(val); err == nil {
				resp.UID = uidVal
			}
			continue
		}
		if strings.HasPrefix(opt, "gid=") {
			val := opt[len("gid="):]
			if gidVal, err := strconv.Atoi(val); err == nil {
				resp.GID = gidVal
			}
			continue
		}
		if resp.ExtraOptions == "" {
			resp.ExtraOptions = opt
		} else {
			resp.ExtraOptions += "," + opt
		}
	}
	return resp
}

func transformVolumeTypeBtrfs(vol *volume.Volume) *VolumeBtrfsOptionsResp {
	return &VolumeBtrfsOptionsResp{
		Device:       vol.Options["device"],
		ExtraOptions: vol.Options["o"],
	}
}
