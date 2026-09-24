package appdeploymentserviceimpl

import (
	"context"

	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/infra/database"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/tasklog"
)

// prepareDockerAPI has every node's agent serve the socket of an app given the
// Docker API before its new task looks for it. The agents would get there at
// their next tick; a task that starts first finds no socket, and many apps give
// up on that and restart. A failure here is the agents', not the deployment's:
// it is logged, and the deployment goes on.
func (s *service) prepareDockerAPI(ctx context.Context, db database.IDB, data *appDeploymentData) error {
	access, err := s.dockerAPIService.AccessOf(ctx, db, data.App.ID)
	if err != nil || access == nil {
		return hperrors.Wrap(err)
	}
	if err = s.dockerAPIService.SyncAgents(ctx); err != nil {
		_ = data.LogStore.Add(ctx, tasklog.NewOutFrame(
			"Not every node serves the app's Docker API yet: "+err.Error(), tasklog.TsNow))
	}
	return nil
}
