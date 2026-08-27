package taskperiodicjobexec

import (
	"context"

	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/infra/database"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/funcutil"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/logging"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/healthcheckservice"
	"github.com/hivepaas/hivepaas/hivepaas_app/tasks/queue"
)

type Executor struct {
	logger logging.Logger
	db     *database.DB

	healthcheckService healthcheckservice.Service
}

func NewExecutor(
	logger logging.Logger,
	db *database.DB,
	taskQueue queue.TaskQueue,

	healthcheckService healthcheckservice.Service,
) *Executor {
	e := &Executor{
		logger: logger,
		db:     db,

		healthcheckService: healthcheckService,
	}
	taskQueue.RegisterPeriodicExecutor(e.execute)
	return e
}

type taskData struct {
	*queue.PeriodicExecData
}

func (e *Executor) execute(
	ctx context.Context,
	execData *queue.PeriodicExecData,
) (err error) {
	data := &taskData{
		PeriodicExecData: execData,
	}
	defer funcutil.EnsureNoPanic(&err)

	switch base.PeriodicKind(data.PeriodicSetting.Kind) {
	case base.PeriodicKindHealthCheck:
		_, err = e.healthcheckService.Healthcheck(ctx, &healthcheckservice.HealthcheckReq{
			PeriodicExecData: execData,
			Healthcheck:      execData.PeriodicSetting.MustAsPeriodicJob().Healthcheck,
		})
		if err != nil {
			return hperrors.Wrap(err)
		}
	default:
	}

	return nil
}
