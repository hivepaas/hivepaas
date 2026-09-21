package appsettingsdto

import (
	"fmt"
	"strings"

	"github.com/moby/moby/api/types/mount"
	"github.com/moby/moby/api/types/swarm"
	vld "github.com/tiendc/go-validator"

	"github.com/hivepaas/hivepaas/hivepaas_app/basedto"
	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/copier"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/fileutil"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/unit"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/volumeservice"
)

type GetAppStorageSettingsReq struct {
	ProjectID    string `json:"-"`
	ProjectEnvID string `json:"-"`
	AppID        string `json:"-"`
}

func NewGetAppStorageSettingsReq() *GetAppStorageSettingsReq {
	return &GetAppStorageSettingsReq{}
}

func (req *GetAppStorageSettingsReq) Validate() hperrors.ValidationErrors {
	validators := make([]vld.Validator, 0, 10) //nolint:mnd
	validators = append(validators, basedto.ValidateID(&req.ProjectID, true, "projectId")...)
	validators = append(validators, basedto.ValidateID(&req.ProjectEnvID, true, "projectEnv")...)
	validators = append(validators, basedto.ValidateID(&req.AppID, true, "appId")...)
	return hperrors.NewValidationErrors(vld.Validate(validators...))
}

type GetAppStorageSettingsResp struct {
	Meta *basedto.Meta        `json:"meta"`
	Data *StorageSettingsResp `json:"data"`
}

type StorageSettingsResp struct {
	Mounts []*Mount `json:"mounts,omitempty"`
	// BorrowedBy is the other side of the same arrangement: the apps that have
	// been given a directory of this app's. It is here because a grant only the
	// borrower can see is a grant nobody will ever revoke.
	BorrowedBy []*MountBorrower `json:"borrowedBy,omitempty"`
	UpdateVer  int              `json:"updateVer"`
}

// MountBorrower is an app that mounts a directory of the app being read.
type MountBorrower struct {
	AppID string `json:"appId"`
	Name  string `json:"name"`
	// Target is where the directory appears inside that app, and Subpath which
	// part of this app's storage it is - empty for the whole of it.
	Target  string `json:"target"`
	Subpath string `json:"subpath,omitempty"`
	Write   bool   `json:"write,omitempty"`
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
	SourceApp      *MountSourceApp   `json:"sourceApp,omitempty"`
}

// MountSourceApp says the directory this mount reaches belongs to another app -
// the files of the database a file manager was put there to work on. Absent is
// the ordinary case: the app's own directory.
//
// Write is stated rather than defaulted. Seeing another app's files is one
// decision and changing them is another, and a request that says nothing has
// only asked for the first.
type MountSourceApp struct {
	AppID string `json:"appId"`
	Write bool   `json:"write,omitempty"`

	// Name and Dangling are answers, not requests: they are filled in when the
	// settings are read and ignored when they are written. Dangling means the
	// directory is still there and the app that owned it is not.
	Name     string `json:"name,omitempty"`
	Dangling bool   `json:"dangling,omitempty"`
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

	// MountDescs says whose directory each mount reaches, index-aligned with the
	// service's mounts. Without it every mount reads as the app's own, which is
	// what every mount was before an app could be given another app's storage.
	MountDescs []*volumeservice.AppMountDesc
	// AppsByKey turns the key found in a directory path back into an app of this
	// environment. A key with no app is a directory whose owner was deleted.
	AppsByKey map[string]*entity.App
}

func TransformStorageSettings(
	input *StorageSettingsTransformInput,
) (resp *StorageSettingsResp, err error) {
	resp = &StorageSettingsResp{
		UpdateVer: int(input.Service.Version.Index), //nolint:gosec
	}

	resp.Mounts, err = TransformStorageMounts(input)
	if err != nil {
		return nil, hperrors.Wrap(err)
	}

	return resp, nil
}

func TransformStorageMounts(
	input *StorageSettingsTransformInput,
) ([]*Mount, error) {
	mounts := input.Service.Spec.TaskTemplate.ContainerSpec.Mounts
	resp := make([]*Mount, 0, len(mounts))
	for i := range mounts {
		var desc *volumeservice.AppMountDesc
		if i < len(input.MountDescs) {
			desc = input.MountDescs[i]
		}
		itemResp, err := TransformStorageMount(&mounts[i], desc, input)
		if err != nil {
			return nil, hperrors.Wrap(err)
		}
		resp = append(resp, itemResp)
	}
	return resp, nil
}

func TransformStorageMount(
	mnt *mount.Mount,
	desc *volumeservice.AppMountDesc,
	input *StorageSettingsTransformInput,
) (resp *Mount, err error) {
	if err = copier.Copy(&resp, mnt); err != nil {
		return nil, hperrors.Wrap(err)
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

	applyMountSourceApp(resp, desc, input)
	return resp, nil
}

// applyMountSourceApp names the app whose directory a mount reaches, when that
// is not the app being read. The subpath is rewritten to what lies below that
// app's directory, so the screen shows the same relative path the owner would
// see for the same place.
func applyMountSourceApp(resp *Mount, desc *volumeservice.AppMountDesc, input *StorageSettingsTransformInput) {
	if desc == nil || desc.AppKey == "" || desc.Own {
		return
	}

	resp.SourceApp = &MountSourceApp{Write: !resp.ReadOnly}
	if owner := input.AppsByKey[desc.AppKey]; owner != nil {
		resp.SourceApp.AppID, resp.SourceApp.Name = owner.ID, owner.Name
	} else {
		// The directory is still there and the app that owned it is not. Saying so
		// is better than showing a path and leaving the reader to guess.
		resp.SourceApp.Name, resp.SourceApp.Dangling = desc.AppKey, true
	}

	if resp.VolumeOptions != nil {
		resp.VolumeOptions.Subpath = desc.Subpath
	}
	if resp.ClusterOptions != nil {
		resp.ClusterOptions.Subpath = desc.Subpath
	}
}

func removeAutoPrefixFromSubpath(subpath string, trimPrefixes []string) string {
	subpath = strings.TrimPrefix(subpath, "/")
	for _, prefix := range trimPrefixes {
		subpath = strings.TrimPrefix(subpath, prefix)
	}
	subpath = strings.TrimPrefix(subpath, "/")
	return subpath
}
