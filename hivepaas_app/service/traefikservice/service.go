package traefikservice

import (
	"context"

	"github.com/moby/moby/api/types/swarm"

	"github.com/hivepaas/hivepaas/hivepaas_app/infra/database"
)

type Service interface {
	GetTraefikSwarmService(ctx context.Context) (*swarm.Service, error)
	RestartTraefikSwarmService(ctx context.Context) error

	ReloadTraefikConfig(ctx context.Context, restartServiceOnFailure bool) error
	ResetTraefikConfig(ctx context.Context) error

	ApplyAppConfig(ctx context.Context, db database.IDB, req *ApplyAppConfigReq) (*ApplyAppConfigResp, error)
	RemoveAppConfig(ctx context.Context, db database.IDB, req *RemoveAppConfigReq) (*RemoveAppConfigResp, error)

	ApplyTrustedIPsToWebEntrypoints(ctx context.Context, req *ApplyTrustedIPsReq) (*ApplyTrustedIPsResp, error)
}
