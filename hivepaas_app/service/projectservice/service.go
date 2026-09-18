package projectservice

import (
	"context"

	"github.com/moby/moby/api/types/swarm"

	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/infra/database"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/bunex"
)

type Service interface {
	LoadProject(ctx context.Context, db database.IDB, projectID string, requireActive bool,
		extraLoadOpts ...bunex.SelectQueryOption) (*entity.Project, error)
	LoadProjects(ctx context.Context, db database.IDB, projectIDs []string, requireActive bool,
		extraLoadOpts ...bunex.SelectQueryOption) ([]*entity.Project, error)

	InitRootProject(ctx context.Context, db database.IDB) (postInitFunc func() error, err error)

	PersistProjectData(ctx context.Context, db database.IDB, data *PersistingProjectData) error
	// DeleteProject removes a project with everything in it. removeStorage also
	// deletes the volumes the project owns and the data its apps kept.
	DeleteProject(ctx context.Context, db database.IDB, project *entity.Project, removeStorage bool) error
	SyncProject(ctx context.Context, db database.IDB, project *entity.Project) (
		newApps, updateApps []*entity.App, _ []swarm.Service, _ error)

	LoadProjectEnv(ctx context.Context, db database.IDB, projectID, projectEnvID string,
		requireProjectActive, requireAppActive bool, extraOpts ...bunex.SelectQueryOption) (
		*entity.ProjectEnv, error)
	// DeleteProjectEnv removes an environment with every app in it. removeStorage
	// also deletes the volumes the environment owns and the data its apps kept.
	DeleteProjectEnv(ctx context.Context, db database.IDB, projectEnv *entity.ProjectEnv,
		removeStorage bool) error
	SetProjectEnvStatus(ctx context.Context, db database.IDB, projectEnv *entity.ProjectEnv,
		status base.ProjectStatus, recursive bool) error

	ExecuteEnvInTx(ctx context.Context, projectEnv *entity.ProjectEnv, requireUpdateVerMatch bool,
		fn func(database.Tx) error) error
}
