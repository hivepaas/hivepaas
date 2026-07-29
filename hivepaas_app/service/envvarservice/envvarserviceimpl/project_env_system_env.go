package envvarserviceimpl

import (
	"context"

	"github.com/hivepaas/hivepaas/hivepaas_app/infra/database"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/envvarservice"
)

func (s *service) BuildSystemEnvVarsInProjectEnv(
	ctx context.Context,
	_ database.IDB,
	_ *envvarservice.BuildSystemEnvVarsInProjectEnvReq,
) ([]*envvarservice.EnvVar, error) {
	return nil, nil
}
