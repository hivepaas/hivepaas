package appsettingsdto

import (
	"strings"

	"github.com/moby/moby/api/types/mount"
	"github.com/moby/moby/api/types/swarm"
	vld "github.com/tiendc/go-validator"

	"github.com/localpaas/localpaas/localpaas_app/apperrors"
	"github.com/localpaas/localpaas/localpaas_app/basedto"
	"github.com/localpaas/localpaas/localpaas_app/entity"
	"github.com/localpaas/localpaas/localpaas_app/pkg/copier"
	"github.com/localpaas/localpaas/localpaas_app/pkg/fileutil"
	"github.com/localpaas/localpaas/localpaas_app/pkg/strutil"
	"github.com/localpaas/localpaas/localpaas_app/pkg/unit"
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
	validators = append(validators, basedto.ValidateID(&req.ProjectID, true, "projectId")...)
	validators = append(validators, basedto.ValidateID(&req.AppID, true, "appId")...)
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
	Type           mount.Type        `json:"type"`
	Target         string            `json:"target"`
	ReadOnly       bool              `json:"readOnly,omitempty"`
	Consistency    mount.Consistency `json:"consistency,omitempty"`
	BindOptions    *BindOptions      `json:"bindOptions,omitempty"`
	VolumeOptions  *VolumeOptions    `json:"volumeOptions,omitempty"`
	TmpfsOptions   *TmpfsOptions     `json:"tmpfsOptions,omitempty"`
	ClusterOptions *ClusterOptions   `json:"clusterOptions,omitempty"`
}

type BindOptions struct {
	BaseDir         string `json:"baseDir"`
	Subpath         string `json:"subpath"`
	SubpathRequired string `json:"subpathRequired"`

	Propagation            mount.Propagation `json:"propagation"`
	NonRecursive           bool              `json:"nonRecursive"`
	CreateMountpoint       bool              `json:"createMountpoint"`
	ReadOnlyNonRecursive   bool              `json:"readOnlyNonRecursive"`
	ReadOnlyForceRecursive bool              `json:"readOnlyForceRecursive"`
}

type VolumeOptions struct {
	Volume          string `json:"volume"`
	Subpath         string `json:"subpath"`
	SubpathRequired string `json:"subpathRequired"`

	NoCopy       bool              `json:"noCopy"`
	Labels       map[string]string `json:"labels"`
	DriverConfig *VolumeDriver     `json:"driverConfig"`
}

type VolumeDriver struct {
	Name    string            `json:"name"`
	Options map[string]string `json:"options"`
}

type TmpfsOptions struct {
	Size    unit.DataSize     `json:"size" copy:"SizeBytes"`
	Mode    fileutil.FileMode `json:"mode"`
	Options [][]string        `json:"options"`
}

type ClusterOptions struct {
	VolumeOptions
}

func TransformStorageSettings(
	app *entity.App,
	storageSettings *entity.StorageSettings,
	service *swarm.Service,
) (resp *StorageSettingsResp, err error) {
	spec := &service.Spec
	resp = &StorageSettingsResp{
		UpdateVer: int(service.Version.Index), //nolint:gosec
	}

	resp.Mounts, err = TransformStorageMounts(app, storageSettings, spec.TaskTemplate.ContainerSpec.Mounts)
	if err != nil {
		return nil, apperrors.Wrap(err)
	}

	return resp, nil
}

func TransformStorageMount(
	app *entity.App,
	storageSettings *entity.StorageSettings,
	mnt *mount.Mount,
) (resp *Mount, err error) {
	if err = copier.Copy(&resp, mnt); err != nil {
		return nil, apperrors.Wrap(err)
	}

	switch mnt.Type { //nolint:exhaustive
	case mount.TypeBind:
		if resp.BindOptions == nil {
			resp.BindOptions = &BindOptions{}
		}
		mntResp := resp.BindOptions
		mntSettings := storageSettings.BindSettings
		if mntSettings == nil {
			mntSettings = &entity.StorageBindSettings{}
		}

		mntResp.SubpathRequired = mntSettings.CaclRequiredSubpath(app)
		baseDir, _, found := strutil.Cut(mnt.Source, mntResp.SubpathRequired)
		if found {
			mntResp.BaseDir = baseDir
			mntResp.Subpath = strings.TrimLeft(strings.TrimPrefix(mnt.Source, baseDir), "/\\")
		} else {
			mntResp.BaseDir = mnt.Source
			mntResp.Subpath = ""
		}

	case mount.TypeVolume:
		if resp.VolumeOptions == nil {
			resp.VolumeOptions = &VolumeOptions{}
		}
		mntResp := resp.VolumeOptions
		mntSettings := storageSettings.VolumeSettings
		if mntSettings == nil {
			mntSettings = &entity.StorageVolumeSettings{}
		}

		mntResp.Volume = mnt.Source
		mntResp.SubpathRequired = mntSettings.CaclRequiredSubpath(app)
		if mnt.VolumeOptions != nil {
			mntResp.Subpath = mnt.VolumeOptions.Subpath
		}

	case mount.TypeCluster:
		if resp.ClusterOptions == nil {
			resp.ClusterOptions = &ClusterOptions{}
		}
		mntResp := resp.ClusterOptions
		mntSettings := storageSettings.ClusterVolumeSettings
		if mntSettings == nil {
			mntSettings = &entity.StorageClusterVolumeSettings{}
		}

		mntResp.Volume = mnt.Source
		mntResp.SubpathRequired = mntSettings.CaclRequiredSubpath(app)
		if mnt.VolumeOptions != nil {
			mntResp.Subpath = mnt.VolumeOptions.Subpath
		}
	}

	return resp, nil
}

func TransformStorageMounts(
	app *entity.App,
	storageSettings *entity.StorageSettings,
	mounts []mount.Mount,
) ([]*Mount, error) {
	resp := make([]*Mount, 0, len(mounts))
	for _, mnt := range mounts {
		itemResp, err := TransformStorageMount(app, storageSettings, &mnt)
		if err != nil {
			return nil, apperrors.Wrap(err)
		}
		resp = append(resp, itemResp)
	}
	return resp, nil
}
