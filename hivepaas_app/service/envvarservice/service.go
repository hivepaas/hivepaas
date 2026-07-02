package envvarservice

import (
	"context"

	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/infra/database"
)

type Service interface {
	HasSecretRef(v string) bool

	BuildAppEnvVars(ctx context.Context, db database.IDB, app *entity.App, buildPhase bool) (
		res []*EnvVar, refSecrets []*entity.Secret, err error)
	ProcessEnvRefs(ctx context.Context, db database.IDB, app *entity.App, envVars []*entity.EnvVar,
		loadEnvVars bool, loadSecrets bool, buildPhase bool) (
		res []*EnvVar, refSecrets []*entity.Secret, err error)
}
