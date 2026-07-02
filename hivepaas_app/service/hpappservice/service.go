package hpappservice

import (
	"context"

	"github.com/moby/moby/api/types/swarm"

	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/infra/database"
)

type Service interface {
	GetHpAppSwarmService(ctx context.Context) (*swarm.Service, error)
	GetHpAppTasks(ctx context.Context) ([]swarm.Task, error)
	RestartHpAppSwarmService(ctx context.Context) error
	ReloadHpAppConfig(ctx context.Context) error

	GetAppReleaseInfo(ctx context.Context) (*AppReleaseInfo, error)
	UpdateSystemVersion(ctx context.Context, db database.IDB, targetVersion *base.ReleaseInfo) error

	GetHpWorkerSwarmService(ctx context.Context) (*swarm.Service, error)
	RestartHpWorkerSwarmService(ctx context.Context) error
	SyncHpWorkerSwarmServiceConfig(mainAppSvc, workerSvc *swarm.Service)

	GetHpUpdaterSwarmService(ctx context.Context) (*swarm.Service, error)
	RestartHpUpdaterSwarmService(ctx context.Context) error
	ShutdownHpUpdaterSwarmService(ctx context.Context) error

	GetHpDbSwarmService(ctx context.Context) (*swarm.Service, error)
	RestartHpDbSwarmService(ctx context.Context) error

	GetHpCacheSwarmService(ctx context.Context) (*swarm.Service, error)
	RestartHpCacheSwarmService(ctx context.Context) error
}
