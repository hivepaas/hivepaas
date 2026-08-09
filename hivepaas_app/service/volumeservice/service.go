package volumeservice

import (
	"context"

	"github.com/moby/moby/api/types/mount"
	"github.com/moby/moby/api/types/volume"
	"github.com/moby/moby/client"

	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/infra/database"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/bunex"
)

type Service interface {
	Rsync(ctx context.Context, source, target *mount.Mount, options ...RsyncOption) error
	EnsureVolumePermissions(ctx context.Context, volMount *mount.Mount, subpaths ...string) error

	MakeSubDirInHost(ctx context.Context, baseDirInHost string, subpath string, requireBaseDirExist bool) error

	CreateProjectDefaultVolume(ctx context.Context, project *entity.Project) (
		*entity.Setting, *client.VolumeCreateResult, error)
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
