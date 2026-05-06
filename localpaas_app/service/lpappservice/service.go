package lpappservice

import (
	"context"

	"github.com/moby/moby/api/types/swarm"

	"github.com/localpaas/localpaas/localpaas_app/base"
	"github.com/localpaas/localpaas/localpaas_app/infra/database"
)

type Service interface {
	GetLpAppSwarmService(ctx context.Context) (*swarm.Service, error)
	GetLpAppTasks(ctx context.Context) ([]swarm.Task, error)
	RestartLpAppSwarmService(ctx context.Context) error
	ReloadLpAppConfig(ctx context.Context) error

	GetAppReleaseInfo(ctx context.Context) (*AppReleaseInfo, error)
	UpdateSystemVersion(ctx context.Context, db database.IDB, targetVersion *base.ReleaseInfo) error

	GetLpWorkerSwarmService(ctx context.Context) (*swarm.Service, error)
	RestartLpWorkerSwarmService(ctx context.Context) error

	GetLpUpdaterSwarmService(ctx context.Context) (*swarm.Service, error)
	RestartLpUpdaterSwarmService(ctx context.Context) error
	ShutdownLpUpdaterSwarmService(ctx context.Context) error

	GetLpDbSwarmService(ctx context.Context) (*swarm.Service, error)
	RestartLpDbSwarmService(ctx context.Context) error

	GetLpCacheSwarmService(ctx context.Context) (*swarm.Service, error)
	RestartLpCacheSwarmService(ctx context.Context) error
}
