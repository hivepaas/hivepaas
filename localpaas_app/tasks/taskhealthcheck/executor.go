package taskhealthcheck

import (
	"context"
	"fmt"
	"time"

	"github.com/tiendc/gofn"

	"github.com/localpaas/localpaas/localpaas_app/apperrors"
	"github.com/localpaas/localpaas/localpaas_app/base"
	"github.com/localpaas/localpaas/localpaas_app/entity"
	"github.com/localpaas/localpaas/localpaas_app/infra/database"
	"github.com/localpaas/localpaas/localpaas_app/infra/logging"
	"github.com/localpaas/localpaas/localpaas_app/infra/rediscache"
	"github.com/localpaas/localpaas/localpaas_app/pkg/funcutil"
	"github.com/localpaas/localpaas/localpaas_app/pkg/timeutil"
	"github.com/localpaas/localpaas/localpaas_app/repository"
	"github.com/localpaas/localpaas/localpaas_app/repository/cacherepository"
	"github.com/localpaas/localpaas/localpaas_app/service/appservice"
	"github.com/localpaas/localpaas/localpaas_app/service/notificationservice"
	"github.com/localpaas/localpaas/localpaas_app/service/settingservice"
	"github.com/localpaas/localpaas/localpaas_app/service/userservice"
	"github.com/localpaas/localpaas/localpaas_app/tasks/queue"
)

type Executor struct {
	logger              logging.Logger
	db                  *database.DB
	redisClient         rediscache.Client
	settingRepo         repository.SettingRepo
	notifEventRepo      cacherepository.HealthcheckNotifEventRepo
	appService          appservice.Service
	settingService      settingservice.Service
	userService         userservice.Service
	notificationService notificationservice.Service
}

func NewExecutor(
	logger logging.Logger,
	db *database.DB,
	taskQueue queue.TaskQueue,
	redisClient rediscache.Client,
	settingRepo repository.SettingRepo,
	notifEventRepo cacherepository.HealthcheckNotifEventRepo,
	appService appservice.Service,
	settingService settingservice.Service,
	userService userservice.Service,
	notificationService notificationservice.Service,
) *Executor {
	e := &Executor{
		logger:              logger,
		db:                  db,
		redisClient:         redisClient,
		settingRepo:         settingRepo,
		notifEventRepo:      notifEventRepo,
		appService:          appService,
		settingService:      settingService,
		userService:         userService,
		notificationService: notificationService,
	}
	taskQueue.RegisterHealthcheckExecutor(e.execute)
	return e
}

type taskData struct {
	*queue.HealthcheckExecData
	Output       *entity.TaskHealthcheckOutput
	NotifMsgData *notificationservice.TemplateDataHealthcheck
}

func (e *Executor) execute(
	ctx context.Context,
	execData *queue.HealthcheckExecData,
) (err error) {
	task := execData.Task
	data := &taskData{
		HealthcheckExecData: execData,
	}
	data.Output = &entity.TaskHealthcheckOutput{}

	var testErr error
	defer func() {
		task.Status = gofn.If(testErr == nil, base.TaskStatusDone, base.TaskStatusFailed)
		task.EndedAt = timeutil.NowUTC()
		task.MustSetOutput(data.Output)

		err = e.sendNotification(ctx, e.db, data)
	}()
	defer funcutil.EnsureNoPanic(&err) // Make sure we catch panic before the above defer

	retries := 0
	startTime := time.Now()
	for {
		switch data.Healthcheck.HealthcheckType {
		case base.HealthcheckTypeREST:
			testErr = e.doHealthcheckREST(ctx, data)
		case base.HealthcheckTypeGRPC:
			testErr = e.doHealthcheckGRPC(ctx, data)
		default:
			testErr = apperrors.NewUnsupported(
				fmt.Sprintf("Healthcheck type '%v'", data.Healthcheck.HealthcheckType))
		}
		if testErr != nil {
			retries++
			if retries > task.Config.MaxRetry {
				break
			}
			task.Config.Retry = retries
			if task.Config.RetryDelay > 0 {
				time.Sleep(task.Config.RetryDelay.ToDuration())
			}
			if time.Since(startTime)+5*time.Second > data.Healthcheck.Interval.ToDuration() {
				break
			}
		} else {
			break
		}
	}

	return err
}
