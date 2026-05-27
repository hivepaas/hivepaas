package queueimpl

import (
	"context"

	"github.com/localpaas/localpaas/localpaas_app/apperrors"
	"github.com/localpaas/localpaas/localpaas_app/base"
	"github.com/localpaas/localpaas/localpaas_app/config"
	"github.com/localpaas/localpaas/localpaas_app/infra/database"
	"github.com/localpaas/localpaas/localpaas_app/infra/gocronqueue"
	"github.com/localpaas/localpaas/localpaas_app/infra/logging"
	"github.com/localpaas/localpaas/localpaas_app/infra/rediscache"
	"github.com/localpaas/localpaas/localpaas_app/repository"
	"github.com/localpaas/localpaas/localpaas_app/repository/cacherepository"
	"github.com/localpaas/localpaas/localpaas_app/service/schedjobservice"
	"github.com/localpaas/localpaas/localpaas_app/service/settingservice"
	"github.com/localpaas/localpaas/localpaas_app/service/startupservice"
	"github.com/localpaas/localpaas/localpaas_app/service/taskservice"
	"github.com/localpaas/localpaas/localpaas_app/tasks/queue"
)

type taskQueue struct {
	db                        *database.DB
	config                    *config.Config
	logger                    logging.Logger
	server                    *gocronqueue.Server
	client                    *gocronqueue.Client
	redisClient               rediscache.Client
	settingRepo               repository.SettingRepo
	taskRepo                  repository.TaskRepo
	taskInfoRepo              cacherepository.TaskInfoRepo
	healthcheckSettingsRepo   cacherepository.HealthcheckSettingsRepo
	healthcheckNotifEventRepo cacherepository.HealthcheckNotifEventRepo
	schedJobService           schedjobservice.Service
	taskService               taskservice.Service
	settingService            settingservice.Service
	startupService            startupservice.Service

	taskExecutorMap     map[base.TaskType]gocronqueue.TaskExecFunc
	healthcheckExecutor queue.HealthcheckExecFunc
}

func New(
	db *database.DB,
	config *config.Config,
	logger logging.Logger,
	redisClient rediscache.Client,
	settingRepo repository.SettingRepo,
	taskRepo repository.TaskRepo,
	cacheTaskInfoRepo cacherepository.TaskInfoRepo,
	healthcheckSettingsRepo cacherepository.HealthcheckSettingsRepo,
	healthcheckNotifEventRepo cacherepository.HealthcheckNotifEventRepo,
	schedJobService schedjobservice.Service,
	taskService taskservice.Service,
	settingService settingservice.Service,
	startupService startupservice.Service,
) queue.TaskQueue {
	return &taskQueue{
		db:                        db,
		config:                    config,
		logger:                    logger,
		redisClient:               redisClient,
		settingRepo:               settingRepo,
		taskRepo:                  taskRepo,
		taskInfoRepo:              cacheTaskInfoRepo,
		healthcheckSettingsRepo:   healthcheckSettingsRepo,
		healthcheckNotifEventRepo: healthcheckNotifEventRepo,
		schedJobService:           schedJobService,
		taskService:               taskService,
		settingService:            settingService,
		startupService:            startupService,
	}
}

func (q *taskQueue) Start() (err error) {
	ctx := context.Background()
	lpSttg, err := q.startupService.LoadLocalPaaSServiceSetting(ctx)
	if err != nil {
		return apperrors.Wrap(err)
	}
	lpSettings := lpSttg.MustAsLocalPaaSService()

	runWorker := q.isWorkerMode()
	if q.isAppMode() {
		runWorker = lpSettings.WorkerSettings.RunWorkerInMainApp
	}

	// Initialize task queue worker if configured
	if runWorker {
		q.logger.Infof("starting task queue worker...")
		q.server, err = gocronqueue.NewServer(&gocronqueue.Config{
			TaskMap:                 q.taskExecutorMap,
			RedisClient:             q.redisClient,
			Logger:                  q.logger,
			Concurrency:             lpSettings.WorkerSettings.Concurrency,
			TaskCheckInterval:       lpSettings.TaskSettings.TaskCheckInterval.ToDuration(),
			TaskCheckFunc:           q.findSchedulingTasks,
			TaskCreateInterval:      lpSettings.TaskSettings.TaskCreateInterval.ToDuration(),
			TaskCreateFunc:          q.doCreateTasksForJobs,
			TaskCanScheduleFunc:     q.canScheduleTask,
			HealthcheckBaseInterval: lpSettings.HealthcheckSettings.BaseInterval.ToDuration(),
			HealthcheckFunc:         q.doHealthcheck,
		})
		if err != nil {
			return apperrors.Wrap(err)
		}

		go func() {
			if err = q.server.Start(); err != nil {
				q.logger.Errorf("failed to start task queue worker: %v", err)
			}
		}()
	}

	// Initialize task queue client (always init a task queue client)
	q.logger.Infof("starting task queue client...")
	q.client, err = gocronqueue.NewClient(q.redisClient, q.logger)
	if err != nil {
		return apperrors.Wrap(err)
	}

	return nil
}

func (q *taskQueue) Shutdown() error {
	q.logger.Info("stopping task queue ...")
	if q.server != nil {
		if err := q.server.Shutdown(); err != nil {
			q.logger.Errorf("failed to start task queue server: %v", err)
			return apperrors.Wrap(err)
		}
	}
	if q.client != nil {
		if err := q.client.Close(); err != nil {
			q.logger.Errorf("failed to stop task queue client: %v", err)
			return apperrors.Wrap(err)
		}
	}
	return nil
}

func (q *taskQueue) StartScheduler() error {
	if q.server == nil {
		q.logger.Error("task queue server is not running")
		return apperrors.New(apperrors.ErrUnavailable).WithParam("Name", "Task queue server")
	}
	if err := q.server.StartScheduler(); err != nil {
		q.logger.Errorf("failed to start scheduler in task queue server: %v", err)
		return apperrors.Wrap(err)
	}
	return nil
}

func (q *taskQueue) StartAllSchedulers() error {
	if q.client != nil {
		if err := q.client.StartScheduler(context.Background()); err != nil {
			q.logger.Errorf("failed to send start scheduler message to servers: %v", err)
			return apperrors.Wrap(err)
		}
	}
	if q.server != nil {
		return q.StartScheduler()
	}
	return nil
}

func (q *taskQueue) StopScheduler() error {
	if q.server == nil {
		q.logger.Error("task queue server is not running")
		return apperrors.New(apperrors.ErrUnavailable).WithParam("Name", "Task queue server")
	}
	if err := q.server.StopScheduler(); err != nil {
		q.logger.Errorf("failed to stop scheduler in task queue server: %v", err)
		return apperrors.Wrap(err)
	}
	return nil
}

func (q *taskQueue) StopAllSchedulers() error {
	if q.client != nil {
		if err := q.client.StopScheduler(context.Background()); err != nil {
			q.logger.Errorf("failed to send stop scheduler message to servers: %v", err)
			return apperrors.Wrap(err)
		}
	}
	if q.server != nil {
		return q.StopScheduler()
	}
	return nil
}

func (q *taskQueue) isAppMode() bool {
	return q.config.RunMode == config.RunModeApp || q.config.RunMode == config.RunModeAppAndWorker
}

func (q *taskQueue) isWorkerMode() bool {
	return q.config.RunMode == config.RunModeWorker || q.config.RunMode == config.RunModeAppAndWorker
}
