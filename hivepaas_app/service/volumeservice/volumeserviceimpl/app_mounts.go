package volumeserviceimpl

import (
	"context"
	"fmt"
	"path/filepath"
	"slices"
	"strings"

	"github.com/moby/moby/api/types/mount"

	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/infra/database"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/bunex"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/entityutil"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/projecthelper"
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

// appScopePrefix is the directory inside a volume of this scope that belongs to
// the app. It is where calcMountSubpath starts from, and what RemoveAppStorage
// measures a directory against before deleting it: one answer, so the two can
// never disagree about what an app owns.
//
// An empty string means there is no answer - the scope gives the app no
// directory of its own, or the app was loaded without what naming one needs.
// Nothing below such a volume is then the app's to claim or to delete, which is
// the safe way round: deleting an app whose storage cannot be identified leaves
// a directory behind, and the other way would take somebody else's with it.
//
// The environment's key is derived from the id rather than read from the
// relation, because the app reaching here is not always loaded with one - the
// deletion path loads the row alone.
func appScopePrefix(app *entity.App, scope base.ObjectScopeType) string {
	envKey := ""
	if app.ProjectEnv != nil {
		envKey = app.ProjectEnv.Key
	} else {
		_, envKey = projecthelper.ParseProjectEnvID(app.ProjectEnvID)
	}

	switch scope {
	case base.ObjectScopeGlobal:
		if app.Project == nil || app.Project.Key == "" || envKey == "" {
			return ""
		}
		return fmt.Sprintf("%v/%v/%v", app.Project.Key, envKey, app.Key)
	case base.ObjectScopeProject:
		if envKey == "" {
			return ""
		}
		return fmt.Sprintf("%v/%v", envKey, app.Key)
	case base.ObjectScopeProjectEnv, base.ObjectScopeApp:
		return app.Key
	case base.ObjectScopeUser, base.ObjectScopeHivepaas:
	}
	return ""
}

// appOwnsSubpath reports whether a directory inside a volume is the app's own:
// its directory, or something below it. A mount built for another app's
// directory answers false, and that is the whole of what stops deleting one app
// from deleting another's data.
func appOwnsSubpath(app *entity.App, scope base.ObjectScopeType, subpath string) bool {
	prefix := appScopePrefix(app, scope)
	if prefix == "" || subpath == "" {
		return false
	}
	subpath = strings.TrimPrefix(filepath.Clean(subpath), "/")
	return subpath == prefix || strings.HasPrefix(subpath, prefix+"/")
}

func calcMountSubpath(
	app *entity.App,
	mnt *volumeservice.AppMountReq,
	setting *entity.Setting,
) string {
	// A mount that names an owner reaches that app's directory instead of the
	// caller's. Everything after this line is the same for both: the request's
	// own subpath is still joined below whichever directory it turned out to be.
	owner := app
	if mnt.OwnerApp != nil {
		owner = mnt.OwnerApp
	}
	subpath := appScopePrefix(owner, setting.Scope)

	if mnt.Type == mount.TypeVolume && mnt.VolumeOptions != nil {
		subpath = filepath.Join(subpath, mnt.VolumeOptions.Subpath)
	}
	if mnt.Type == mount.TypeCluster && mnt.ClusterOptions != nil {
		subpath = filepath.Join(subpath, mnt.ClusterOptions.Subpath)
	}
	return subpath
}
