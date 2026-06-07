package projectservice

import (
	"context"

	"github.com/moby/moby/api/types/swarm"

	"github.com/localpaas/localpaas/localpaas_app/entity"
	"github.com/localpaas/localpaas/localpaas_app/infra/database"
	"github.com/localpaas/localpaas/localpaas_app/pkg/bunex"
)

type Service interface {
	LoadProject(ctx context.Context, db database.IDB, projectID string, requireActive bool,
		extraLoadOpts ...bunex.SelectQueryOption) (*entity.Project, error)
	LoadProjects(ctx context.Context, db database.IDB, projectIDs []string, requireActive bool,
		extraLoadOpts ...bunex.SelectQueryOption) ([]*entity.Project, error)

	InitRootProject(ctx context.Context, db database.IDB) (postInitFunc func() error, err error)

	PersistProjectData(ctx context.Context, db database.IDB, data *PersistingProjectData) error
	DeleteProject(ctx context.Context, db database.IDB, project *entity.Project) error
	SyncProject(ctx context.Context, db database.IDB, project *entity.Project) (
		newApps, updateApps []*entity.App, _ []swarm.Service, _ error)
}
