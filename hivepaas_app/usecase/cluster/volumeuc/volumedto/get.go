package volumedto

import (
	"strconv"
	"strings"
	"time"

	"github.com/moby/moby/api/types/mount"
	vld "github.com/tiendc/go-validator"
	"github.com/tiendc/gofn"

	"github.com/hivepaas/hivepaas/hivepaas_app/basedto"
	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
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

func (req *GetVolumeReq) Validate() hperrors.ValidationErrors {
	validators := make([]vld.Validator, 0, 10) //nolint:mnd
	validators = append(validators, req.GetSettingReq.Validate()...)
	return hperrors.NewValidationErrors(vld.Validate(validators...))
}

type GetVolumeResp struct {
	Meta *basedto.Meta `json:"meta"`
	Data *VolumeResp   `json:"data"`
}

type VolumeResp struct {
	*settings.BaseSettingResp

	// Pinned volume node
	NodeID    string `json:"nodeId,omitempty"`
	NodeLabel string `json:"nodeLabel,omitempty"`

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
		return nil, hperrors.Wrap(err)
	}

	resp.BaseSettingResp, err = settings.TransformSettingBase(setting)
	if err != nil {
		return nil, hperrors.Wrap(err)
	}

	// The setting says how the volume was specified. Until docker has actually
	// built it, that specification is the best answer there is - better than the
	// blank driver a freshly created, not-yet-mounted volume would show otherwise.
	if volEnt.Driver != "" {
		resp.Driver = docker.VolumeDriver(volEnt.Driver)
		resp.Options = volEnt.DriverOpts
		resp.Labels = volEnt.Labels
		if typed := transformVolumeTypeOptions(resp, resp.Options); typed {
			resp.Options = nil
		}
	}

	// Docker is the authority on a volume that actually exists, so where the two
	// disagree it wins outright - and it alone can supply Mountpoint, Scope,
	// CreatedAt, RefCount and Size, which the setting never records. Every field
	// docker can answer is reassigned unconditionally, so nothing from the
	// setting survives under a docker volume that turns out to disagree with it.
	vol := refClusterObjects.RefVolumes[setting.RefID]
	if vol != nil {
		resp.Driver = docker.VolumeDriver(vol.Driver)
		resp.Mountpoint = vol.Mountpoint
		resp.Options = vol.Options
		resp.Scope = docker.VolumeScope(vol.Scope)
		resp.Labels = vol.Labels
		resp.CreatedAt = transformVolumeCreatedAt(vol.CreatedAt)
		if vol.UsageData != nil {
			resp.RefCount = vol.UsageData.RefCount
			resp.Size = vol.UsageData.Size
		}

		if typed := transformVolumeTypeOptions(resp, resp.Options); typed {
			resp.Options = nil
		}
	}

	return resp, nil
}

// transformVolumeTypeOptions breaks driver options for docker's `local` driver
// into the typed block their `type` selects and reports whether one matched -
// the same parsing whether opts came from the recorded specification or from a
// live docker volume, and any earlier typed block is cleared either way so a
// stale one never survives a source that turns out not to have one.
func transformVolumeTypeOptions(resp *VolumeResp, opts map[string]string) (typed bool) {
	resp.BindOptions, resp.NfsOptions, resp.TmpfsOptions = nil, nil, nil
	if resp.Driver != docker.VolumeDriverLocal {
		return false
	}
	switch opts["type"] {
	case "none":
		resp.BindOptions = transformVolumeTypeBind(opts)
	case "nfs":
		resp.NfsOptions = transformVolumeTypeNfs(opts)
	case "tmpfs":
		resp.TmpfsOptions = transformVolumeTypeTmpfs(opts)
	default:
		return false
	}
	return true
}

func transformVolumeCreatedAt(createdAt string) time.Time {
	t, err := time.Parse(time.RFC3339, createdAt)
	if err == nil {
		return t
	}
	return time.Time{}
}

func transformVolumeTypeBind(
	opts map[string]string,
) *VolumeBindOptionsResp {
	optList := strings.Split(opts["o"], ",")
	if !gofn.Contain(optList, "bind") {
		return nil
	}
	resp := &VolumeBindOptionsResp{
		Directory: opts["device"],
	}
	for _, opt := range optList {
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

func transformVolumeTypeNfs(opts map[string]string) *VolumeNfsOptionsResp {
	resp := &VolumeNfsOptionsResp{
		Device: opts["device"],
	}
	optList := strings.Split(opts["o"], ",")
	for _, opt := range optList {
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

func transformVolumeTypeTmpfs(opts map[string]string) *VolumeTmpfsOptionsResp {
	resp := &VolumeTmpfsOptionsResp{
		Device: opts["device"],
	}
	optList := strings.Split(opts["o"], ",")
	for _, opt := range optList {
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
