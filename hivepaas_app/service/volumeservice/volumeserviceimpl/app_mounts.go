package volumeserviceimpl

import (
	"context"
	"fmt"
	"path/filepath"
	"slices"

	"github.com/moby/moby/api/types/mount"

	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/infra/database"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/bunex"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/entityutil"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/volumeservice"
)

func (s *service) BuildAppMounts(
	ctx context.Context,
	db database.IDB,
	req *volumeservice.BuildAppMountsReq,
) (*volumeservice.BuildAppMountsResp, error) {
	app := req.App

	volumeIDs := make([]string, 0, len(req.New))
	for _, mnt := range req.New {
		// For custom mounts, only support type Volume and Cluster
		if mnt.Type != mount.TypeVolume && mnt.Type != mount.TypeCluster {
			return nil, hperrors.Wrap(hperrors.ErrUnsupported).
				WithParam("Name", fmt.Sprintf("Mount type '%v'", mnt.Type))
		}
		volumeIDs = append(volumeIDs, mnt.Source)
	}

	// Validate volumes can be used by the project
	volumes, err := s.settingRepo.ListByIDs(ctx, db, app.GetObjectScope(), volumeIDs, true,
		bunex.SelectWhere("setting.type = ?", base.SettingTypeClusterVolume),
	)
	if err != nil {
		return nil, hperrors.Wrap(err)
	}
	volumeMap := entityutil.SliceToIDMap(volumes)

	mounts := slices.Clone(req.Kept)
	for _, mnt := range req.New {
		setting, found := volumeMap[mnt.Source]
		if !found {
			return nil, hperrors.NewNotFound("Volume").WithMsgLog("volume %v not found", mnt.Source)
		}
		dockerMnt, err := s.buildAppMount(ctx, app, mnt, setting)
		if err != nil {
			return nil, hperrors.Wrap(err)
		}
		mounts = append(mounts, *dockerMnt)
	}

	// Every volume visible to the app's scope, not just the requested ones: a
	// pinned volume behind an unchanged bind mount carries no id to look it up
	// by, so resolving the mounts back to their pins needs the same candidate set
	// placementserviceimpl matches mounts against.
	scopeVolumes, _, err := s.settingRepo.List(ctx, db, app.GetObjectScope(), nil,
		bunex.SelectWhere("setting.type = ?", base.SettingTypeClusterVolume),
	)
	if err != nil {
		return nil, hperrors.Wrap(err)
	}
	// This is the last point before the mounts are saved where a contradiction
	// can be caught: applying a contradictory pin set silently drops the
	// placement constraint instead of failing.
	if err = refuseConflictingVolumePins(mounts, scopeVolumes); err != nil {
		return nil, hperrors.Wrap(err)
	}

	return &volumeservice.BuildAppMountsResp{Mounts: mounts}, nil
}

func (s *service) buildAppMount(
	ctx context.Context,
	app *entity.App,
	mnt *volumeservice.AppMountReq,
	setting *entity.Setting,
) (*mount.Mount, error) {
	vol, err := setting.AsClusterVolume()
	if err != nil {
		return nil, hperrors.Wrap(err)
	}
	dockerMnt := &mount.Mount{
		Type:        mnt.Type,
		Source:      setting.RefID,
		Target:      mnt.Target,
		ReadOnly:    mnt.ReadOnly,
		Consistency: mnt.Consistency,
	}

	s.buildDockerMount(ctx, dockerMnt, mnt, vol, setting, app)

	// Ensure full permissions (0777) on all mounted volume subpaths before starting/updating the service
	if dockerMnt.Type == mount.TypeVolume || dockerMnt.Type == mount.TypeCluster {
		subpath := ""
		if dockerMnt.VolumeOptions != nil {
			subpath = dockerMnt.VolumeOptions.Subpath
		}
		_ = s.ensureVolumePermissions(ctx, dockerMnt, subpath)
	}
	return dockerMnt, nil
}

func (s *service) buildDockerMount(
	ctx context.Context,
	dockerMnt *mount.Mount,
	mnt *volumeservice.AppMountReq,
	vol *entity.ClusterVolume,
	setting *entity.Setting,
	app *entity.App,
) {
	subpath := calcMountSubpath(app, mnt, setting)
	s.useBindMountIfAppropriate(ctx, dockerMnt, vol, subpath)

	switch dockerMnt.Type {
	case mount.TypeVolume:
		if opts := mnt.VolumeOptions; opts != nil {
			dockerMnt.VolumeOptions = &mount.VolumeOptions{
				Subpath:      subpath,
				NoCopy:       opts.NoCopy,
				Labels:       opts.Labels,
				DriverConfig: opts.DriverConfig,
			}
		}
		applyVolumeDriverConfigUnlessOverridden(dockerMnt, vol)
	case mount.TypeCluster:
		if opts := mnt.ClusterOptions; opts != nil {
			dockerMnt.VolumeOptions = &mount.VolumeOptions{
				Subpath:      subpath,
				NoCopy:       opts.NoCopy,
				Labels:       opts.Labels,
				DriverConfig: opts.DriverConfig,
			}
		}
	case mount.TypeBind, mount.TypeImage, mount.TypeTmpfs, mount.TypeNamedPipe:
	}
}

func (s *service) useBindMountIfAppropriate(
	ctx context.Context,
	dockerMnt *mount.Mount,
	vol *entity.ClusterVolume,
	subpath string,
) {
	directory, propagation, ok := bindMountTarget(vol, subpath)
	if !ok {
		return
	}
	if err := s.makeSubDirInHost(ctx, vol.DriverOpts["device"], subpath, true); err != nil {
		return
	}

	dockerMnt.Type = mount.TypeBind
	dockerMnt.Source = directory
	dockerMnt.BindOptions = &mount.BindOptions{
		// Kept for volumes with no pin, which claim every node reaches the same
		// data: creating the directory there is the right thing. A pinned volume
		// is kept on its node by a placement constraint instead.
		CreateMountpoint: true,
	}
	if propagation != "" {
		dockerMnt.BindOptions.Propagation = propagation
	}
	// Reset all other kind of options
	dockerMnt.VolumeOptions = nil
	dockerMnt.ClusterOptions = nil
	dockerMnt.TmpfsOptions = nil
	dockerMnt.ImageOptions = nil
}

func calcMountSubpath(
	app *entity.App,
	mnt *volumeservice.AppMountReq,
	setting *entity.Setting,
) string {
	var subpath string
	switch setting.Scope {
	case base.ObjectScopeGlobal:
		subpath = fmt.Sprintf("%v/%v/%v", app.Project.Key, app.ProjectEnv.Key, app.Key)
	case base.ObjectScopeProject:
		subpath = fmt.Sprintf("%v/%v", app.ProjectEnv.Key, app.Key)
	case base.ObjectScopeProjectEnv, base.ObjectScopeApp:
		subpath = app.Key
	case base.ObjectScopeUser, base.ObjectScopeHivepaas:
	}

	if mnt.Type == mount.TypeVolume && mnt.VolumeOptions != nil {
		subpath = filepath.Join(subpath, mnt.VolumeOptions.Subpath)
	}
	if mnt.Type == mount.TypeCluster && mnt.ClusterOptions != nil {
		subpath = filepath.Join(subpath, mnt.ClusterOptions.Subpath)
	}
	return subpath
}
