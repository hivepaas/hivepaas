package traefikservice

import (
	"context"

	"github.com/moby/moby/api/types/swarm"

	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
)

type Service interface {
	GetTraefikSwarmService(ctx context.Context) (*swarm.Service, error)
	RestartTraefikSwarmService(ctx context.Context) error

	ReloadTraefikConfig(ctx context.Context, restartServiceOnFailure bool) error
	ResetTraefikConfig(ctx context.Context) error

	ApplyAppConfig(ctx context.Context, app *entity.App, service *swarm.Service, data *AppConfigData) error
	RemoveAppConfig(ctx context.Context, app *entity.App, service *swarm.Service) error
}
