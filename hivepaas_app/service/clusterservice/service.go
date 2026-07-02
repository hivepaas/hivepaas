package clusterservice

import (
	"context"
	"time"

	"github.com/moby/moby/api/types/swarm"
	"github.com/moby/moby/client"

	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/infra/database"
	"github.com/hivepaas/hivepaas/services/docker"
)

type Service interface {
	PersistClusterData(ctx context.Context, db database.IDB, data *PersistingClusterData) error

	IsMultiNode(ctx context.Context) (bool, error)

	// Docker services
	ServiceInspect(ctx context.Context, serviceID string, caching bool) (*swarm.Service, error)
	ServiceUpdate(ctx context.Context, serviceID string, version *swarm.Version, service *swarm.ServiceSpec,
		options ...docker.ServiceUpdateOption) (*client.ServiceUpdateResult, error)
	ServiceRemove(ctx context.Context, serviceID string, retryMax int, retryDelay time.Duration) error
	ServicesRemove(ctx context.Context, serviceIDs []string, retryMax int, retryDelay time.Duration) error

	// Docker secrets
	CreateSecretForApp(ctx context.Context, db database.IDB, app *entity.App, secret *entity.Secret) (
		*entity.SwarmSecretRef, error)
	CreateSecretsForApp(ctx context.Context, db database.IDB, app *entity.App, secrets []*entity.Secret) (
		[]*entity.SwarmSecretRef, error)
	UpdateSecretForApp(ctx context.Context, db database.IDB, app *entity.App, oldSecret, newSecret *entity.Secret) error
	DeleteSecretForApp(ctx context.Context, db database.IDB, app *entity.App, secret *entity.Secret) error
	SecretRemove(ctx context.Context, secretID string, retryMax int, retryDelay time.Duration) error
	SecretsRemove(ctx context.Context, secretIDs []string, retryMax int, retryDelay time.Duration) error

	// Docker config
	CreateConfigForApp(ctx context.Context, db database.IDB, app *entity.App, secret *entity.ConfigFile) (
		*entity.SwarmConfigRef, error)
	CreateConfigsForApp(ctx context.Context, db database.IDB, app *entity.App, configs []*entity.ConfigFile) (
		[]*entity.SwarmConfigRef, error)
	UpdateConfigForApp(ctx context.Context, db database.IDB, app *entity.App,
		oldSecret, newSecret *entity.ConfigFile) error
	DeleteConfigForApp(ctx context.Context, db database.IDB, app *entity.App, secret *entity.ConfigFile) error
	ConfigRemove(ctx context.Context, configID string, retryMax int, retryDelay time.Duration) error
	ConfigsRemove(ctx context.Context, configIDs []string, retryMax int, retryDelay time.Duration) error
}
