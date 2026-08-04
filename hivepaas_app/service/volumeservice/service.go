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
	Rsync(
		ctx context.Context,
		source, target *mount.Mount,
		options ...RsyncOption,
	) error

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
}
