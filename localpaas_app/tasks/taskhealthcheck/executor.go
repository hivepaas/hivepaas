package taskhealthcheck

import (
	"context"

	"github.com/localpaas/localpaas/localpaas_app/apperrors"
	"github.com/localpaas/localpaas/localpaas_app/infra/database"
	"github.com/localpaas/localpaas/localpaas_app/infra/logging"
	"github.com/localpaas/localpaas/localpaas_app/pkg/funcutil"
	"github.com/localpaas/localpaas/localpaas_app/repository/cacherepository"
	"github.com/localpaas/localpaas/localpaas_app/service/healthcheckservice"
	"github.com/localpaas/localpaas/localpaas_app/service/notificationservice"
	"github.com/localpaas/localpaas/localpaas_app/tasks/queue"
)

type Executor struct {
	logger logging.Logger
	db     *database.DB

	notifEventRepo cacherepository.HealthcheckNotifEventRepo

	healthcheckService  healthcheckservice.Service
	notificationService notificationservice.Service
}

func NewExecutor(
	logger logging.Logger,
	db *database.DB,
	taskQueue queue.TaskQueue,

	notifEventRepo cacherepository.HealthcheckNotifEventRepo,

	healthcheckService healthcheckservice.Service,
	notificationService notificationservice.Service,
) *Executor {
	e := &Executor{
		logger:              logger,
		db:                  db,
		notifEventRepo:      notifEventRepo,
		healthcheckService:  healthcheckService,
		notificationService: notificationService,
	}
	taskQueue.RegisterHealthcheckExecutor(e.execute)
	return e
}

type taskData struct {
	*queue.HealthcheckExecData
	NotifMsgData *notificationservice.TemplateDataHealthcheck
}

func (e *Executor) execute(
	ctx context.Context,
	execData *queue.HealthcheckExecData,
) (err error) {
	data := &taskData{
		HealthcheckExecData: execData,
	}

	defer func() {
		err = e.sendNotification(ctx, e.db, data)
	}()
	defer funcutil.EnsureNoPanic(&err) // Make sure we catch panic before the above defer

	_, err = e.healthcheckService.Healthcheck(ctx, &healthcheckservice.HealthcheckReq{
		HealthcheckExecData: execData,
	})
	if err != nil {
		return apperrors.New(err)
	}
	return nil
}
