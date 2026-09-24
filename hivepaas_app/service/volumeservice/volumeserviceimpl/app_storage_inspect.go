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

	// Grouped by volume, because one volume is opened once however many apps sit
	// on it: a template of eight apps on the project's volume is one directory
	// read, or one command on the node that holds it - not eight of either.
	states := make([]*volumeservice.AppStorageState, 0, len(req.Queries))
	perVolume := map[string][]*volumeservice.AppStorageState{}
	order := make([]string, 0, len(byID))
	for _, query := range req.Queries {
		state := &volumeservice.AppStorageState{AppKey: query.AppKey, VolumeID: query.VolumeID}
		states = append(states, state)

		setting := byID[query.VolumeID]
		if setting == nil || query.App == nil {
			continue
		}
		state.VolumeName = setting.Name
		if state.Path = appStoragePath(query.App, setting.Scope, query.Subpath); state.Path == "" {
			continue
		}
		if _, seen := perVolume[setting.ID]; !seen {
			order = append(order, setting.ID)
		}
		perVolume[setting.ID] = append(perVolume[setting.ID], state)
	}

	for _, volumeID := range order {
		s.inspectVolume(ctx, byID[volumeID], perVolume[volumeID])
	}
	return &volumeservice.InspectAppStorageResp{States: states}, nil
}

// inspectVolume fills in one volume's states: by reading the directories here
// when they can be reached, and by asking the node that holds them when they
// cannot.
//
// Neither being possible leaves every state unchecked, which is the honest
// answer - nothing was seen, which is not the same as nothing being there.
func (s *service) inspectVolume(
	ctx context.Context,
	setting *entity.Setting,
	states []*volumeservice.AppStorageState,
) {
	root := s.storageRootOnThisNode(ctx, setting)
	if root == "" {
		s.inspectVolumeOnItsNode(ctx, setting, states)
		return
	}
	for _, state := range states {
		entries, err := os.ReadDir(filepath.Join(root, state.Path))
		if err != nil {
			// A directory that is not there is the ordinary case: this app has
			// never run. Anything else is a directory that exists and cannot be
			// read, which is not something to warn about either.
			state.Checked = true
			continue
		}
		state.Checked, state.Exists = true, true
		state.Empty = len(entries) == 0
	}
}

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
		where := &appStorage{constraint: constraint, local: s.storageIsOnThisNode(ctx, pins)}
		if err = s.removeStorageTarget(ctx, &target, where); err != nil {
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
