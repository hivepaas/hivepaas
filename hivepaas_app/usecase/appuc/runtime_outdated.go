package appuc

import (
	"context"

	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/bunex"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/functionservice/functionbuild"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/systemappservice"
)

// functionRuntimeOutdated says whether the deployment a function runs was built
// on other runtime images than this HivePaaS version's. A function never
// deployed, or deployed before deployments recorded their runtime images, is
// not: nothing says it is.
func (uc *UC) functionRuntimeOutdated(ctx context.Context, appID string) (bool, error) {
	deployments, _, err := uc.deploymentRepo.List(ctx, uc.db, appID, nil, latestDoneDeployment()...)
	if err != nil {
		return false, hperrors.Wrap(err)
	}
	if len(deployments) == 0 || deployments[0].Output == nil {
		return false, nil
	}
	return functionbuild.RuntimeOutdated(deployments[0].Output.RuntimeImages,
		systemappservice.CurrentRelease().FunctionRuntimes), nil
}

// latestDoneDeployment selects an app's latest deployment that succeeded: the
// one it runs, since a failed or canceled one changed nothing.
func latestDoneDeployment() []bunex.SelectQueryOption {
	return []bunex.SelectQueryOption{
		bunex.SelectColumns("deployment.id", "deployment.output"),
		bunex.SelectWhere("deployment.status = ?", base.DeploymentStatusDone),
		bunex.SelectOrder("deployment.created_at DESC"),
		bunex.SelectLimit(1),
	}
}
