package volumeserviceimpl

import (
	"context"
	"path/filepath"
	"strings"

	"github.com/moby/moby/api/types/mount"

	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/infra/database"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/bunex"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/volumeservice"
)

func (s *service) DescribeAppMounts(
	ctx context.Context,
	db database.IDB,
	app *entity.App,
	mounts []mount.Mount,
) ([]*volumeservice.AppMountDesc, error) {
	// The same candidate set the mounts were built from, and the same one
	// deleting the app matches them against: a bind carries no volume id, so the
	// volume it came from is found by the directory it points into.
	volumes, _, err := s.settingRepo.List(ctx, db, app.GetObjectScope(), nil,
		bunex.SelectWhere("setting.type = ?", base.SettingTypeClusterVolume),
	)
	if err != nil {
		return nil, hperrors.Wrap(err)
	}

	out := make([]*volumeservice.AppMountDesc, len(mounts))
	for i := range mounts {
		out[i] = describeAppMount(app, &mounts[i], volumes)
	}
	return out, nil
}

// describeAppMount reads a mount back into the app whose directory it reaches.
//
// appStorageTarget is what resolves it: written for deletion, it answers the
// question asked here as well - which volume this mount is in, and which
// directory inside it - and using the same one keeps the screen's account of a
// mount and the deletion's account of it from ever disagreeing.
func describeAppMount(
	app *entity.App,
	mnt *mount.Mount,
	volumes []*entity.Setting,
) *volumeservice.AppMountDesc {
	desc := &volumeservice.AppMountDesc{}

	target, ok := appStorageTarget(mnt, volumes)
	if !ok || target.volume == nil {
		return desc // a volume mounted whole, or one nothing here accounts for
	}
	scope := target.volume.Scope

	if appOwnsSubpath(app, scope, target.subpath) {
		desc.AppKey, desc.Own = app.Key, true
		desc.Subpath = trimDirPrefix(target.subpath, appScopePrefix(app, scope))
		return desc
	}

	key, rest, ok := appKeyInSubpath(scope, target.subpath)
	if !ok {
		return desc
	}
	desc.AppKey, desc.Subpath = key, rest
	return desc
}

// appKeyInSubpath pulls an app's key out of a directory inside a volume of this
// scope, and returns what is left below it. The depth is the one appScopePrefix
// writes, read the other way round.
func appKeyInSubpath(scope base.ObjectScopeType, subpath string) (key, rest string, ok bool) {
	var depth int
	switch scope {
	case base.ObjectScopeGlobal:
		depth = 3 // <project>/<env>/<app>
	case base.ObjectScopeProject:
		depth = 2 // <env>/<app>
	case base.ObjectScopeProjectEnv, base.ObjectScopeApp:
		depth = 1 // <app>
	case base.ObjectScopeUser, base.ObjectScopeHivepaas:
		return "", "", false
	}

	parts := strings.Split(strings.TrimPrefix(filepath.Clean(subpath), "/"), "/")
	if len(parts) < depth || parts[depth-1] == "" {
		return "", "", false
	}
	return parts[depth-1], strings.Join(parts[depth:], "/"), true
}

func trimDirPrefix(path, prefix string) string {
	return strings.TrimPrefix(strings.TrimPrefix(path, prefix), "/")
}
