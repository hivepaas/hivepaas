package appsettingsdto

import (
	"strings"

	"github.com/moby/moby/api/types/mount"
	"github.com/moby/moby/api/types/swarm"
	"github.com/moby/moby/api/types/volume"
	vld "github.com/tiendc/go-validator"

	"github.com/hivepaas/hivepaas/hivepaas_app/apperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/basedto"
	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/copier"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/fileutil"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/strutil"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/unit"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/settings/storagesettingsuc/storagesettingsdto"
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
	Settings *storagesettingsdto.StorageSettingsResp `json:"settings"`
	Mounts   []*Mount                                `json:"mounts,omitempty"`

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

type StorageSettingsTransformInput struct {
	App             *entity.App
	Project         *entity.Project
	Setting         *entity.Setting
	Service         *swarm.Service
	ReturningMounts []*mount.Mount
	Volumes         []*volume.Volume
}

func TransformStorageSettings(
	input *StorageSettingsTransformInput,
) (resp *StorageSettingsResp, err error) {
	resp = &StorageSettingsResp{
		UpdateVer: int(input.Service.Version.Index), //nolint:gosec
	}

	resp.Settings, err = storagesettingsdto.TransformStorageSettings(
		&storagesettingsdto.StorageSettingsTransformInput{
			Project: input.Project,
			App:     input.App,
			Setting: input.Setting,
			Volumes: input.Volumes,
		})
	if err != nil {
		return nil, apperrors.Wrap(err)
	}

	resp.Mounts, err = TransformStorageMounts(input, resp.Settings)
	if err != nil {
		return nil, apperrors.Wrap(err)
	}

	return resp, nil
}

func TransformStorageMounts(
	input *StorageSettingsTransformInput,
	storageSettings *storagesettingsdto.StorageSettingsResp,
) ([]*Mount, error) {
	resp := make([]*Mount, 0, len(input.ReturningMounts))
	for _, mnt := range input.ReturningMounts {
		itemResp, err := TransformStorageMount(storageSettings, mnt)
		if err != nil {
			return nil, apperrors.Wrap(err)
		}
		resp = append(resp, itemResp)
	}
	return resp, nil
}

func TransformStorageMount(
	storageSettings *storagesettingsdto.StorageSettingsResp,
	mnt *mount.Mount,
) (resp *Mount, err error) {
	if err = copier.Copy(&resp, mnt); err != nil {
		return nil, apperrors.Wrap(err)
	}

	switch mnt.Type {
	case mount.TypeBind:
		if resp.BindOptions == nil {
			resp.BindOptions = &BindOptions{}
		}
		mntResp := resp.BindOptions
		mntResp.SubpathRequired = storageSettings.BindSettings.SubpathRequired
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
		mntResp.Volume = mnt.Source
		mntResp.SubpathRequired = storageSettings.VolumeSettings.SubpathRequired
		if mnt.VolumeOptions != nil {
			mntResp.Subpath = mnt.VolumeOptions.Subpath
		}

	case mount.TypeCluster:
		if resp.ClusterOptions == nil {
			resp.ClusterOptions = &ClusterOptions{}
		}
		mntResp := resp.ClusterOptions
		mntResp.Volume = mnt.Source
		mntResp.SubpathRequired = storageSettings.ClusterVolumeSettings.SubpathRequired
		if mnt.VolumeOptions != nil {
			mntResp.Subpath = mnt.VolumeOptions.Subpath
		}

	case mount.TypeTmpfs, mount.TypeNamedPipe, mount.TypeImage:
		// Do nothing
	}

	return resp, nil
}
