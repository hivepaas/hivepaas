package systemappserviceimpl

import (
	"context"
	"fmt"
	"time"

	"github.com/moby/moby/api/types/swarm"
	"github.com/moby/moby/client"

	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/infra/database"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/bunex"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/timeutil"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/systemappservice"
)

// deployingWithin is how recent a deployment not yet done must be to count as
// one being made. One older was left behind - a worker that stopped mid-way -
// and must not keep the app from being looked at for good.
const deployingWithin = time.Hour

func (s *service) Check(
	ctx context.Context,
	db database.IDB,
	app *entity.App,
) (*systemappservice.AppCheck, error) {
	deploying, err := s.deploying(ctx, db, app)
	if err != nil {
		return nil, hperrors.Wrap(err)
	}
	if deploying {
		return &systemappservice.AppCheck{Action: entity.SystemAppSyncSkipped, Problem: "it is being deployed"}, nil
	}

	svc, err := s.listedService(ctx, app)
	if err != nil {
		return nil, hperrors.Wrap(err)
	}
	if svc == nil {
		return &systemappservice.AppCheck{Action: entity.SystemAppSyncReported,
			Problem: "its service is gone", ServiceGone: true}, nil
	}
	return checkService(svc), nil
}

// checkService judges a service that exists, by its task counts, which only a
// listing carries.
func checkService(svc *swarm.Service) *systemappservice.AppCheck {
	if svc.UpdateStatus != nil && (svc.UpdateStatus.State == swarm.UpdateStateUpdating ||
		svc.UpdateStatus.State == swarm.UpdateStateRollbackStarted) {
		return &systemappservice.AppCheck{Action: entity.SystemAppSyncSkipped,
			Problem: "its service is being updated"}
	}
	status := svc.ServiceStatus
	switch {
	case status == nil:
		return &systemappservice.AppCheck{Action: entity.SystemAppSyncNone}
	case status.DesiredTasks == 0 && svc.Spec.Mode.Global != nil:
		return &systemappservice.AppCheck{Action: entity.SystemAppSyncReported,
			Problem: "no node can run it"}
	case status.DesiredTasks == 0:
		return &systemappservice.AppCheck{Action: entity.SystemAppSyncReported,
			Problem: "its service is scaled to zero"}
	case status.RunningTasks < status.DesiredTasks:
		return &systemappservice.AppCheck{Action: entity.SystemAppSyncReported,
			Problem: fmt.Sprintf("%d of its %d tasks are running", status.RunningTasks, status.DesiredTasks)}
	}
	return &systemappservice.AppCheck{Action: entity.SystemAppSyncNone}
}

// deploying says whether a deployment of the app is being made.
func (s *service) deploying(ctx context.Context, db database.IDB, app *entity.App) (bool, error) {
	deployments, _, err := s.deploymentRepo.List(ctx, db, app.ID, nil,
		bunex.SelectColumns("deployment.id"),
		bunex.SelectWhereIn("deployment.status IN (?)",
			base.DeploymentStatusNotStarted, base.DeploymentStatusInProgress),
		bunex.SelectWhere("deployment.created_at > ?", timeutil.NowUTC().Add(-deployingWithin)),
		bunex.SelectLimit(1),
	)
	if err != nil {
		return false, hperrors.Wrap(err)
	}
	return len(deployments) > 0, nil
}

// listedService is the app's service with its task counts, nil when it has
// none.
func (s *service) listedService(ctx context.Context, app *entity.App) (*swarm.Service, error) {
	if app.ServiceID == "" {
		return nil, nil
	}
	listed, err := s.dockerManager.ServiceListByIDs(ctx, []string{app.ServiceID},
		func(opts *client.ServiceListOptions) { opts.Status = true })
	if err != nil {
		return nil, hperrors.Wrap(err)
	}
	if len(listed.Items) == 0 {
		return nil, nil
	}
	return &listed.Items[0], nil
}
