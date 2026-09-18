package volumeservice

import (
	"context"
	"time"

	"github.com/moby/moby/api/types/mount"
	"github.com/moby/moby/api/types/volume"

	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/infra/database"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/bunex"
)

type Service interface {
	Rsync(ctx context.Context, source, target *mount.Mount, options ...RsyncOption) error
	EnsureVolumePermissions(ctx context.Context, volMount *mount.Mount, subpaths ...string) error

	// BuildAppMounts turns requested mounts of cluster-volume settings into the
	// docker mounts an app's service carries: it checks every volume is usable
	// from the app's scope, derives each subpath, rewrites a bind volume into a
	// bind mount, fills in driver config, opens up permissions, and refuses a set
	// of mounts that pins the service to more than one node. It does not write the
	// mounts to the service.
	BuildAppMounts(ctx context.Context, db database.IDB, req *BuildAppMountsReq) (*BuildAppMountsResp, error)

	MakeSubDirInHost(ctx context.Context, baseDirInHost string, subpath string, requireBaseDirExist bool) error

	// RemoveAppStorage deletes the directories an app kept its data in, inside the
	// volumes it had mounted. The volumes themselves stay: one volume holds a
	// directory per app, and the other apps are still using theirs.
	//
	// A mount carrying no subpath is left alone. That is the whole volume, which
	// belongs to whoever created it.
	RemoveAppStorage(ctx context.Context, mounts []mount.Mount) error

	RemoveVolume(ctx context.Context, volumeID string, force bool, retryMax int, retryDelay time.Duration) error

	CreateProjectDefaultVolume(ctx context.Context, project *entity.Project) (*entity.Setting, error)
	ListProjectVolumes(ctx context.Context, db database.IDB, project *entity.Project,
		extraOpts ...bunex.SelectQueryOption) ([]*entity.Setting, map[string]*volume.Volume, error)
	RemoveAllProjectVolumes(ctx context.Context, db database.IDB, project *entity.Project,
		force bool) error

	ListProjectEnvVolumes(ctx context.Context, db database.IDB, projectEnv *entity.ProjectEnv,
		extraOpts ...bunex.SelectQueryOption) ([]*entity.Setting, map[string]*volume.Volume, error)
	RemoveAllProjectEnvVolumes(ctx context.Context, db database.IDB, projectEnv *entity.ProjectEnv,
		force bool) error

	// Sync volumes from Docker to app DB
	SyncVolumes(ctx context.Context, db database.IDB) ([]volume.Volume, error)
}
