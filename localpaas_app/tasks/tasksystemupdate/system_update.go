package tasksystemupdate

import (
	"context"

	"github.com/localpaas/localpaas/localpaas_app/apperrors"
	"github.com/localpaas/localpaas/localpaas_app/infra/database"
)

func (e *Executor) updateSystemVersion(
	ctx context.Context,
	_ database.IDB,
	data *taskData,
) (err error) {
	// 1. Pull all images we need
	// err := e.pullAllImages(ctx, data)
	// if err != nil {
	//	return apperrors.Wrap(err)
	// }

	// 2. Scale down the main app to zero instance
	err = e.scaleMainAppService(ctx, 0, data)
	if err != nil {
		return apperrors.Wrap(err)
	}

	// Scale down the workers to zero instance
	err = e.scaleWorkerService(ctx, 0, data)
	if err != nil {
		return apperrors.Wrap(err)
	}

	// TODO: backup DB data

	// 3. Migrate DB schema
	err = e.migrateDBSchema(ctx, data)
	if err != nil {
		return apperrors.Wrap(err)
	}

	// 4. Update redis
	err = e.updateRedisService(ctx, data)
	if err != nil {
		return apperrors.Wrap(err)
	}

	// 5. Update traefik
	err = e.updateTraefikService(ctx, data)
	if err != nil {
		return apperrors.Wrap(err)
	}

	// 6. Update main app then bring it back
	err = e.updateMainAppService(ctx, data)
	if err != nil {
		return apperrors.Wrap(err)
	}

	// 7. Update worker then bring it back
	err = e.updateWorkerService(ctx, data)
	if err != nil {
		return apperrors.Wrap(err)
	}

	return nil
}
