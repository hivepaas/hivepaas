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

	// DescribeAppMounts says whose directory inside a volume each of an app's
	// mounts reaches. It is the reverse of BuildAppMounts: what the service spec
	// carries is a path, and this reads the app out of it again, so the settings
	// screen can say "the files of postgres" where a path would say nothing.
	//
	// The answer is index-aligned with the mounts given.
	DescribeAppMounts(ctx context.Context, db database.IDB, app *entity.App,
		mounts []mount.Mount) ([]*AppMountDesc, error)

	MakeSubDirInHost(ctx context.Context, baseDirInHost string, subpath string, requireBaseDirExist bool) error

	// RemoveAppStorage deletes the directories an app kept its data in, inside the
	// volumes it had mounted. The volumes themselves stay: one volume holds a
	// directory per app, and the other apps are still using theirs.
	//
	// A mount carrying nothing of the app's own is left alone: the whole volume,
	// which belongs to whoever created it, and a bind that is not a managed
	// volume's directory, which HivePaaS did not make.
	//
	// The mounts are matched against the app's volume settings, so the deletion
	// happens on the node the data is pinned to rather than wherever this process
	// runs.
	RemoveAppStorage(ctx context.Context, db database.IDB, app *entity.App, mounts []mount.Mount) error

	// ResetAppStoragePermissions gives what is in the directory of one of an app's
	// mounts to a user, or opens it up to every user - for data the app has to be
	// given that another user wrote: a different image on the same volume, files
	// copied in by hand. Nothing does this on its own; see MakeDirWritableCmd.
	//
	// Only a directory of the app's own is reset. A mount reaching another app's
	// directory, a whole volume or a bind HivePaaS did not make is refused.
	ResetAppStoragePermissions(ctx context.Context, db database.IDB, req *ResetAppStoragePermissionsReq) (
		*ResetAppStoragePermissionsResp, error)

	// InspectAppStorage reports whether the directories apps would be given
	// already hold something. It reads only, and answers about an app's own
	// directory inside a volume rather than about the volume.
	InspectAppStorage(ctx context.Context, db database.IDB, req *InspectAppStorageReq) (
		*InspectAppStorageResp, error)

	// RemoveAppStoragePaths deletes the directories InspectAppStorage reported
	// on. It takes the same queries, because the apps it is asked about are the
	// ones that do not exist yet.
	RemoveAppStoragePaths(ctx context.Context, db database.IDB, req *InspectAppStorageReq) error

	RemoveVolume(ctx context.Context, volumeID string, force bool, retryMax int, retryDelay time.Duration) error

	// RemoveVolumeInCluster removes a volume on the node its data is on.
	//
	// A volume HivePaaS creates lives on the node its pin names and nowhere else,
	// so a removal issued against the manager's daemon usually finds nothing and
	// reports success while the volume stays where it is. This one goes to the
	// right node, through its agent.
	//
	// removeData asks for the data to go as well. It only means anything for a
	// volume made of a host directory, which keeps every byte when the volume
	// itself is removed; for any other volume docker's removal is the data's
	// removal and there is nothing left to ask for.
	RemoveVolumeInCluster(ctx context.Context, setting *entity.Setting, removeData bool,
		force bool, retryMax int, retryDelay time.Duration) error

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
