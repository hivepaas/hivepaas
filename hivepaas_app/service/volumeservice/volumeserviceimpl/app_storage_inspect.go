package volumeserviceimpl

import (
	"context"
	"os"
	"path/filepath"

	"github.com/moby/moby/api/types/mount"

	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/infra/database"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/bunex"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/placementservice"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/volumeservice"
)

// InspectAppStorage reports whether the directories apps would be given already
// hold something.
//
// It answers about the app's own directory inside a volume, never the volume: a
// volume shared by a project is almost always non-empty, and an answer about it
// would be true on every app anybody ever creates. The app need not exist yet -
// what the directory is called comes from the keys, which are known before
// anything is written.
func (s *service) InspectAppStorage(
	ctx context.Context,
	db database.IDB,
	req *volumeservice.InspectAppStorageReq,
) (*volumeservice.InspectAppStorageResp, error) {
	if req == nil || len(req.Queries) == 0 {
		return &volumeservice.InspectAppStorageResp{}, nil
	}

	volumes, _, err := s.settingRepo.List(ctx, db, req.Scope, nil,
		bunex.SelectWhere("setting.type = ?", base.SettingTypeClusterVolume),
	)
	if err != nil {
		return nil, hperrors.Wrap(err)
	}
	byID := make(map[string]*entity.Setting, len(volumes))
	for _, vol := range volumes {
		byID[vol.ID] = vol
	}

	// One volume is opened once however many apps sit on it: a template of eight
	// apps on the project's volume is one directory read, not eight.
	roots := map[string]string{}
	states := make([]*volumeservice.AppStorageState, 0, len(req.Queries))
	for _, query := range req.Queries {
		states = append(states, s.inspectOne(ctx, query, byID, roots))
	}
	return &volumeservice.InspectAppStorageResp{States: states}, nil
}

func (s *service) inspectOne(
	ctx context.Context,
	query *volumeservice.AppStorageQuery,
	byID map[string]*entity.Setting,
	roots map[string]string,
) *volumeservice.AppStorageState {
	state := &volumeservice.AppStorageState{AppKey: query.AppKey, VolumeID: query.VolumeID}

	setting := byID[query.VolumeID]
	if setting == nil || query.App == nil {
		return state
	}
	state.VolumeName = setting.Name

	path := appStoragePath(query.App, setting.Scope, query.Subpath)
	if path == "" {
		return state
	}
	state.Path = path

	root, found := roots[setting.ID]
	if !found {
		root = s.storageRootOnThisNode(ctx, setting)
		roots[setting.ID] = root
	}
	if root == "" {
		// The data is on a volume this node cannot open - pinned elsewhere, or a
		// driver with nothing on a filesystem here. Saying nothing is the honest
		// answer; a guess would be a warning nobody could act on.
		return state
	}

	entries, err := os.ReadDir(filepath.Join(root, path))
	if err != nil {
		// A directory that is not there is the ordinary case: this app has never
		// run. Anything else is a directory that exists and cannot be read, which
		// is not something to warn about either.
		state.Checked = true
		return state
	}
	state.Checked, state.Exists = true, true
	state.Empty = len(entries) == 0
	return state
}

// RemoveAppStoragePaths deletes the directories InspectAppStorage reported on.
//
// It takes the same queries rather than an app, because the apps it is asked
// about do not exist: this is what clears what a previous install left behind,
// in the moment between someone confirming it and the new apps being created.
func (s *service) RemoveAppStoragePaths(
	ctx context.Context,
	db database.IDB,
	req *volumeservice.InspectAppStorageReq,
) error {
	if req == nil || len(req.Queries) == 0 {
		return nil
	}

	volumes, _, err := s.settingRepo.List(ctx, db, req.Scope, nil,
		bunex.SelectWhere("setting.type = ?", base.SettingTypeClusterVolume),
	)
	if err != nil {
		return hperrors.Wrap(err)
	}
	byID := make(map[string]*entity.Setting, len(volumes))
	for _, vol := range volumes {
		byID[vol.ID] = vol
	}

	for _, query := range req.Queries {
		setting := byID[query.VolumeID]
		if setting == nil || query.App == nil {
			continue
		}
		path := appStoragePath(query.App, setting.Scope, query.Subpath)
		if path == "" {
			continue
		}
		target, ok := wholeVolumeTarget(setting, path)
		if !ok {
			continue
		}

		pins, err := placementservice.VolumePinsForMounts([]mount.Mount{target.mount}, volumes)
		if err != nil {
			return hperrors.Wrap(err)
		}
		constraint, conflict := placementservice.VolumePinConstraint(pins)
		if conflict != nil {
			return hperrors.Wrap(hperrors.ErrActionFailed).WithMsgLog(
				"cannot remove storage at %s: %s", path, conflict.Error())
		}
		if err = s.removeStorageTarget(ctx, &target, constraint, s.storageIsOnThisNode(ctx, pins)); err != nil {
			return hperrors.Wrap(err)
		}
	}
	return nil
}

// wholeVolumeTarget is the volume mounted whole, with the directory to delete
// named inside it - the shape removeStorageTarget works on.
func wholeVolumeTarget(setting *entity.Setting, path string) (storageTarget, bool) {
	vol, err := setting.AsClusterVolume()
	if err != nil || vol == nil {
		return storageTarget{}, false
	}
	if directory, _, ok := bindMountTarget(vol, ""); ok {
		return storageTarget{mount: bindMountWhole(directory), subpath: path, volume: setting}, true
	}
	if setting.RefID == "" {
		return storageTarget{}, false
	}
	return storageTarget{
		mount:   mount.Mount{Type: mount.TypeVolume, Source: setting.RefID, Target: volumeHelperTarget},
		subpath: path,
		volume:  setting,
	}, true
}

// storageRootOnThisNode is where a volume's contents can be read from this
// process, or empty when they cannot.
func (s *service) storageRootOnThisNode(ctx context.Context, setting *entity.Setting) string {
	vol, err := setting.AsClusterVolume()
	if err != nil || vol == nil {
		return ""
	}

	mnt := mount.Mount{Type: mount.TypeVolume, Source: setting.RefID}
	if directory, _, ok := bindMountTarget(vol, ""); ok {
		mnt = mount.Mount{Type: mount.TypeBind, Source: directory}
	}
	root, _ := s.getDirectHostPath(ctx, &mnt, "")
	return root
}

// appStoragePath is the directory an app is given inside a volume: the same one
// calcMountSubpath builds when the mount is made, and the same one
// appOwnsSubpath tests when the app is deleted. All three have to agree - a
// check that looks somewhere else reports on a stranger's data, and reports the
// app's own as clean.
func appStoragePath(app *entity.App, scope base.ObjectScopeType, subpath string) string {
	prefix := appScopePrefix(app, scope)
	if prefix == "" {
		return ""
	}
	if subpath == "" {
		return safeSubpath(prefix)
	}
	// The mount's own subpath is judged before it is joined: filepath.Join
	// resolves "../../etc" away first, and what comes out the other side is a
	// directory beside the app's that reads as a plain relative path.
	below := safeSubpath(subpath)
	if below == "" {
		return ""
	}
	return safeSubpath(filepath.Join(prefix, below))
}
