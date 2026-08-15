package hpappservice

import (
	"context"

	"github.com/moby/moby/api/types/swarm"

	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/infra/database"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/bunex"
)

type Service interface {
	LoadAppByKey(ctx context.Context, db database.IDB, appKey string, extraOpts ...bunex.SelectQueryOption) (
		*entity.App, error)

	GetHpAppSwarmService(ctx context.Context) (*swarm.Service, error)
	GetHpAppTasks(ctx context.Context) ([]swarm.Task, error)
	RestartHpAppSwarmService(ctx context.Context) error
	ReloadHpAppConfig(ctx context.Context) error
	SetupHttpSettingsDefault(httpSettings *entity.AppHttpSettings)

	GetAppReleaseInfo(ctx context.Context) (*AppReleaseInfo, error)
	UpdateSystemVersion(ctx context.Context, db database.IDB, targetVersion *base.ReleaseInfo) error

	GetHpWorkerSwarmService(ctx context.Context) (*swarm.Service, error)
	RestartHpWorkerSwarmService(ctx context.Context) error
	SyncHpWorkerSwarmServiceConfig(mainAppSvc, workerSvc *swarm.Service)

	GetHpUpdaterSwarmService(ctx context.Context) (*swarm.Service, error)
	RestartHpUpdaterSwarmService(ctx context.Context) error
	ShutdownHpUpdaterSwarmService(ctx context.Context) error

	GetHpAgentSwarmService(ctx context.Context) (*swarm.Service, error)
	RestartHpAgentSwarmService(ctx context.Context) error
	GetHpAgentImage(ctx context.Context) string

	GetHpDbSwarmService(ctx context.Context) (*swarm.Service, error)
	RestartHpDbSwarmService(ctx context.Context) error

	GetHpCacheSwarmService(ctx context.Context) (*swarm.Service, error)
	RestartHpCacheSwarmService(ctx context.Context) error
}
