package clustersecretservice

import (
	"context"
	"time"

	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/infra/database"
)

type Service interface {
	// Docker secrets
	CreateSecretForApp(ctx context.Context, db database.IDB, app *entity.App,
		secret *entity.Secret) (*entity.SwarmSecretRef, error)
	CreateSecretsForApp(ctx context.Context, db database.IDB, app *entity.App,
		secrets []*entity.Secret) ([]*entity.SwarmSecretRef, error)
	UpdateSecretForApp(ctx context.Context, db database.IDB, app *entity.App,
		oldSecret, newSecret *entity.Secret) error
	UpdateSecretsForApp(ctx context.Context, db database.IDB, app *entity.App,
		oldSecrets, newSecrets []*entity.Secret) error // slices can include `nil` indicating deletion/creation only
	RemoveSecretForApp(ctx context.Context, db database.IDB, app *entity.App,
		secrets ...*entity.Secret) error

	SecretRemove(ctx context.Context, secretID string, retryMax int, retryDelay time.Duration) error
	SecretsRemove(ctx context.Context, secretIDs []string, retryMax int, retryDelay time.Duration) error
	HasSecretChanges(newSecret, oldSecret *entity.Secret) bool

	// Docker config
	CreateConfigForApp(ctx context.Context, db database.IDB, app *entity.App,
		config *entity.ConfigFile) (*entity.SwarmConfigRef, error)
	CreateConfigsForApp(ctx context.Context, db database.IDB, app *entity.App,
		configs []*entity.ConfigFile) ([]*entity.SwarmConfigRef, error)
	UpdateConfigForApp(ctx context.Context, db database.IDB, app *entity.App,
		oldConfig, newConfig *entity.ConfigFile) error
	UpdateConfigsForApp(ctx context.Context, db database.IDB, app *entity.App,
		oldConfigs, newConfigs []*entity.ConfigFile) error // slices can include `nil` indicating deletion/creation only
	RemoveConfigForApp(ctx context.Context, db database.IDB, app *entity.App,
		configs ...*entity.ConfigFile) error

	ConfigRemove(ctx context.Context, configID string, retryMax int, retryDelay time.Duration) error
	ConfigsRemove(ctx context.Context, configIDs []string, retryMax int, retryDelay time.Duration) error
	HasConfigChanges(newConfig, oldConfig *entity.ConfigFile) bool
}
