package taskperiodicjobexec

import (
	"context"

	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/infra/database"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/logging"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/safego"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/functionautoscaleservice"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/healthcheckservice"
	"github.com/hivepaas/hivepaas/hivepaas_app/tasks/queue"
)

type Executor struct {
	logger logging.Logger
	db     *database.DB

	healthcheckService       healthcheckservice.Service
	functionAutoscaleService functionautoscaleservice.Service
}

func NewExecutor(
	logger logging.Logger,
	db *database.DB,
	taskQueue queue.TaskQueue,

	healthcheckService healthcheckservice.Service,
	functionAutoscaleService functionautoscaleservice.Service,
) *Executor {
	e := &Executor{
		logger: logger,
		db:     db,

		healthcheckService:       healthcheckService,
		functionAutoscaleService: functionAutoscaleService,
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
	defer safego.RecoverTo(&err)

	switch base.PeriodicKind(data.PeriodicSetting.Kind) {
	case base.PeriodicKindHealthCheck:
		_, err = e.healthcheckService.Healthcheck(ctx, &healthcheckservice.HealthcheckReq{
			PeriodicExecData: execData,
			Healthcheck:      execData.PeriodicSetting.MustAsPeriodicJob().Healthcheck,
		})
		if err != nil {
			return hperrors.Wrap(err)
		}
	case base.PeriodicKindFunctionAutoscale:
		if err = e.functionAutoscaleService.Run(ctx, execData); err != nil {
			return hperrors.Wrap(err)
		}
	default:
	}

	return nil
}
