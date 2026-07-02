package appdeploymentserviceimpl

import (
	"context"
	"fmt"

	"github.com/hivepaas/hivepaas/hivepaas_app/apperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/infra/database"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/bunex"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/tasklog"
)

const (
	deploymentLockKeyForDockerService = "deployment:docker:service:%v"
)

func (s *service) lockDockerServiceForDeployment(
	ctx context.Context,
	db database.Tx,
	data *appDeploymentData,
) (shouldContinue bool, err error) {
	// Put a lock record in DB for acquiring after (must use separate `db`, not this transaction)
	lockKey := fmt.Sprintf(deploymentLockKeyForDockerService, data.App.ServiceID)
	_ = s.lockRepo.Insert(ctx, s.db, &entity.Lock{ID: lockKey})

	// Acquire the lock
	_, err = s.lockRepo.GetByID(ctx, db, lockKey, bunex.SelectFor("UPDATE"))
	if err != nil {
		_ = data.LogStore.Add(ctx, tasklog.NewErrFrame("failed to create a lock for the app service",
			tasklog.TsNow))
		return false, apperrors.New(err)
	}

	// Now, we have the lock, need to check either this deployment should continue or stop.
	// If there is any "newer" deployment done, need to skip this one.

	newerDeployments, _, err := s.deploymentRepo.List(ctx, s.db, data.App.ID, nil,
		bunex.SelectWhere("deployment.status = ?", base.DeploymentStatusDone),
		bunex.SelectWhere("deployment.created_at >= ?", data.Deployment.CreatedAt),
		bunex.SelectLimit(1),
		bunex.SelectColumns("id"),
	)
	if err != nil {
		return false, apperrors.New(err)
	}

	return len(newerDeployments) == 0, nil
}
