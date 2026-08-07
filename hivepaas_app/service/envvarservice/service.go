package envvarservice

import (
	"context"

	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/infra/database"
)

type Service interface {
	HasRef(v string) bool
	HasSecretRef(v string) bool

	BuildEnvVarsInProject(ctx context.Context, db database.IDB, req *BuildEnvVarsInProjectReq) (
		*BuildEnvVarsInProjectResp, error)
	BuildSystemEnvVarsInProject(ctx context.Context, db database.IDB, req *BuildSystemEnvVarsInProjectReq) (
		[]*EnvVar, error)

	BuildEnvVarsInProjectEnv(ctx context.Context, db database.IDB, req *BuildEnvVarsInProjectEnvReq) (
		*BuildEnvVarsInProjectEnvResp, error)
	BuildSystemEnvVarsInProjectEnv(ctx context.Context, db database.IDB, req *BuildSystemEnvVarsInProjectEnvReq) (
		[]*EnvVar, error)

	BuildEnvVarsInApp(ctx context.Context, db database.IDB, req *BuildEnvVarsInAppReq) (
		*BuildEnvVarsInAppResp, error)
	BuildSharedEnvVarsInApp(ctx context.Context, db database.IDB, app *entity.App, options EnvBuildOptions) (
		[]*EnvVar, error)
	BuildSystemEnvVarsInApp(ctx context.Context, db database.IDB, req *BuildSystemEnvVarsInAppReq) (
		[]*EnvVar, error)

	BuildEnvVarsForAllAppsInScope(ctx context.Context, db database.IDB, scope *entity.ObjectScope,
		buildPhase bool, onlyApps []string, transaction bool, concurrency bool) (
		appEnvVarData []*AppEnvVarData, err error)
	ApplyEnvVarsForApps(ctx context.Context, db database.IDB, appEnvData []*AppEnvVarData,
		transaction bool, concurrency bool) map[int]error
}
