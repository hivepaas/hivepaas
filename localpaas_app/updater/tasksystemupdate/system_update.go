package tasksystemupdate

import (
	"context"

	"github.com/localpaas/localpaas/localpaas_app/apperrors"
	"github.com/localpaas/localpaas/localpaas_app/infra/database"
	"github.com/localpaas/localpaas/localpaas_app/pkg/funcutil"
)

func (e *Executor) stopServices(
	ctx context.Context,
	data *taskData,
) (err error) {
	// 1. Scale down the main app to zero instance
	err = e.scaleMainAppService(ctx, 0, data)
	if err != nil {
		return apperrors.Wrap(err)
	}

	// 2. Scale down the workers to zero instance
	err = e.scaleWorkerService(ctx, 0, data)
	if err != nil {
		return apperrors.Wrap(err)
	}

	return nil
}

func (e *Executor) onBeforeSystemUpdate(
	_ context.Context,
	_ *taskData,
) (err error) {
	// 1. Pull all images we need
	// err = e.pullAllImages(ctx, data)
	// if err != nil {
	//	return apperrors.Wrap(err)
	// }

	// TODO: backup DB data

	return nil
}

func (e *Executor) onAfterSystemUpdate(
	ctx context.Context,
	data *taskData,
) (err error) {
	// Bring back the main app instances
	if data.CurrentAppReplicas != nil && *data.CurrentAppReplicas > 0 {
		err = e.scaleMainAppService(ctx, *data.CurrentAppReplicas, data)
		if err != nil {
			return apperrors.Wrap(err)
		}
	}

	// Bring back the worker instances
	if data.CurrentWorkerReplicas != nil && *data.CurrentWorkerReplicas > 0 {
		err = e.scaleWorkerService(ctx, *data.CurrentWorkerReplicas, data)
		if err != nil {
			return apperrors.Wrap(err)
		}
	}

	return nil
}

func (e *Executor) updateSystem(
	ctx context.Context,
	db database.IDB,
	data *taskData,
) (err error) {
	defer funcutil.EnsureNoPanic(&err)

	// 1. Update DB
	err = e.updateDbService(ctx, db, data)
	if err != nil {
		return apperrors.Wrap(err)
	}

	// 2. Update redis
	err = e.updateRedisService(ctx, data)
	if err != nil {
		return apperrors.Wrap(err)
	}

	// 3. Update traefik
	err = e.updateTraefikService(ctx, data)
	if err != nil {
		return apperrors.Wrap(err)
	}

	// 4. Update main app then bring it back
	err = e.updateMainAppService(ctx, data)
	if err != nil {
		return apperrors.Wrap(err)
	}

	// 5. Update worker then bring it back
	err = e.updateWorkerService(ctx, data)
	if err != nil {
		return apperrors.Wrap(err)
	}

	return nil
}
