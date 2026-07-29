package envvarservice

import (
	"context"

	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/infra/database"
)

type Service interface {
	HasRef(v string) bool
	HasSecretRef(v string) bool

	ComputeEnvVarsInProject(ctx context.Context, db database.IDB, req *ComputeEnvVarsInProjectReq) (
		*ComputeEnvVarsInProjectResp, error)
	ComputeSystemEnvVarsInProject(ctx context.Context, db database.IDB, req *ComputeSystemEnvVarsInProjectReq) (
		[]*EnvVar, error)

	ComputeEnvVarsInProjectEnv(ctx context.Context, db database.IDB, req *ComputeEnvVarsInProjectEnvReq) (
		*ComputeEnvVarsInProjectEnvResp, error)
	ComputeSystemEnvVarsInProjectEnv(ctx context.Context, db database.IDB, req *ComputeSystemEnvVarsInProjectEnvReq) (
		[]*EnvVar, error)

	ComputeEnvVarsInApp(ctx context.Context, db database.IDB, req *ComputeEnvVarsInAppReq) (
		*ComputeEnvVarsInAppResp, error)
	ComputeSharedEnvVarsInApp(ctx context.Context, db database.IDB, app *entity.App, buildPhase bool,
		skipLoadingSecrets bool, maskSecrets bool) ([]*EnvVar, error)
	ComputeSystemEnvVarsInApp(ctx context.Context, db database.IDB, req *ComputeSystemEnvVarsInAppReq) (
		[]*EnvVar, error)
}
