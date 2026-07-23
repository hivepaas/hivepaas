package envvarservice

import (
	"context"

	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/infra/database"
)

type Service interface {
	HasRef(v string) bool
	HasSecretRef(v string) bool

	ComputeProjectEnvVars(ctx context.Context, db database.IDB, req *ComputeProjectEnvVarsReq) (
		[]*EnvVar, error)
	ComputeProjectSystemEnvVars(ctx context.Context, db database.IDB, req *ComputeProjectSystemEnvVarsReq) (
		[]*EnvVar, error)

	ComputeAppEnvVars(ctx context.Context, db database.IDB, req *ComputeAppEnvVarsReq) (
		[]*EnvVar, error)
	ComputeAppSharedEnvVars(ctx context.Context, db database.IDB, app *entity.App, buildPhase bool,
		skipLoadingSecrets bool, maskSecrets bool) ([]*EnvVar, error)
	ComputeAppSystemEnvVars(ctx context.Context, db database.IDB, req *ComputeAppSystemEnvVarsReq) (
		[]*EnvVar, error)
}
