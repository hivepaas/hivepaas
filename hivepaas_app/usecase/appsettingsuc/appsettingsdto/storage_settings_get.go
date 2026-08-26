package appsettingsdto

import (
	"fmt"
	"strings"

	"github.com/moby/moby/api/types/mount"
	"github.com/moby/moby/api/types/swarm"
	vld "github.com/tiendc/go-validator"

	"github.com/hivepaas/hivepaas/hivepaas_app/apperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/basedto"
	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/copier"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/fileutil"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/unit"
)

type GetAppStorageSettingsReq struct {
	ProjectID    string `json:"-"`
	ProjectEnvID string `json:"-"`
	AppID        string `json:"-"`
}

func NewGetAppStorageSettingsReq() *GetAppStorageSettingsReq {
	return &GetAppStorageSettingsReq{}
}

func (req *GetAppStorageSettingsReq) Validate() apperrors.ValidationErrors {
	validators := make([]vld.Validator, 0, 10) //nolint:mnd
	validators = append(validators, basedto.ValidateID(&req.ProjectID, true, "projectId")...)
	validators = append(validators, basedto.ValidateID(&req.ProjectEnvID, true, "projectEnv")...)
	validators = append(validators, basedto.ValidateID(&req.AppID, true, "appId")...)
	return apperrors.NewValidationErrors(vld.Validate(validators...))
}

type GetAppStorageSettingsResp struct {
	Meta *basedto.Meta        `json:"meta"`
	Data *StorageSettingsResp `json:"data"`
}

type StorageSettingsResp struct {
	Mounts    []*Mount `json:"mounts,omitempty"`
	UpdateVer int      `json:"updateVer"`
}

type Mount struct {
	Key            string            `json:"key"`
	Type           mount.Type        `json:"type"`
	Source         string            `json:"source"`
	Target         string            `json:"target"`
	ReadOnly       bool              `json:"readOnly,omitempty"`
	Consistency    mount.Consistency `json:"consistency,omitempty"`
	BindOptions    *BindOptions      `json:"bindOptions,omitempty"`
	VolumeOptions  *VolumeOptions    `json:"volumeOptions,omitempty"`
	ClusterOptions *ClusterOptions   `json:"clusterOptions,omitempty"`
	TmpfsOptions   *TmpfsOptions     `json:"tmpfsOptions,omitempty"`
}

type BindOptions struct {
	Propagation            mount.Propagation `json:"propagation"`
	NonRecursive           bool              `json:"nonRecursive"`
	CreateMountpoint       bool              `json:"createMountpoint"`
	ReadOnlyNonRecursive   bool              `json:"readOnlyNonRecursive"`
	ReadOnlyForceRecursive bool              `json:"readOnlyForceRecursive"`
}

type VolumeOptions struct {
	Subpath      string            `json:"subpath"`
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
	App                *entity.App
	Service            *swarm.Service
	MountKeyCalculator func(*mount.Mount) string
}

func TransformStorageSettings(
	input *StorageSettingsTransformInput,
) (resp *StorageSettingsResp, err error) {
	resp = &StorageSettingsResp{
		UpdateVer: int(input.Service.Version.Index), //nolint:gosec
	}

	resp.Mounts, err = TransformStorageMounts(input)
	if err != nil {
		return nil, apperrors.Wrap(err)
	}

	return resp, nil
}

func TransformStorageMounts(
	input *StorageSettingsTransformInput,
) ([]*Mount, error) {
	mounts := input.Service.Spec.TaskTemplate.ContainerSpec.Mounts
	resp := make([]*Mount, 0, len(mounts))
	for i := range mounts {
		itemResp, err := TransformStorageMount(&mounts[i], input)
		if err != nil {
			return nil, apperrors.Wrap(err)
		}
		resp = append(resp, itemResp)
	}
	return resp, nil
}

func TransformStorageMount(
	mnt *mount.Mount,
	input *StorageSettingsTransformInput,
) (resp *Mount, err error) {
	if err = copier.Copy(&resp, mnt); err != nil {
		return nil, apperrors.Wrap(err)
	}
	resp.Key = input.MountKeyCalculator(mnt)

	app := input.App
	trimSubpathPrefixes := []string{
		fmt.Sprintf("%v/%v/%v", app.Project.Key, app.ProjectEnv.Key, app.Key),
		fmt.Sprintf("%v/%v", app.ProjectEnv.Key, app.Key),
	}

	switch mnt.Type {
	case mount.TypeVolume:
		if resp.VolumeOptions == nil {
			resp.VolumeOptions = &VolumeOptions{}
		}
		if mnt.VolumeOptions != nil {
			resp.VolumeOptions.Subpath = removeAutoPrefixFromSubpath(mnt.VolumeOptions.Subpath, trimSubpathPrefixes)
		}
	case mount.TypeCluster:
		if resp.ClusterOptions == nil {
			resp.ClusterOptions = &ClusterOptions{}
		}
		if mnt.VolumeOptions != nil {
			resp.ClusterOptions.Subpath = removeAutoPrefixFromSubpath(mnt.VolumeOptions.Subpath, trimSubpathPrefixes)
		}
	case mount.TypeBind, mount.TypeTmpfs, mount.TypeNamedPipe, mount.TypeImage:
		// Do nothing
	}

	return resp, nil
}

func removeAutoPrefixFromSubpath(subpath string, trimPrefixes []string) string {
	subpath = strings.TrimPrefix(subpath, "/")
	for _, prefix := range trimPrefixes {
		subpath = strings.TrimPrefix(subpath, prefix)
	}
	subpath = strings.TrimPrefix(subpath, "/")
	return subpath
}
