package queueimpl

import (
	"github.com/localpaas/localpaas/localpaas_app/apperrors"
	"github.com/localpaas/localpaas/localpaas_app/base"
	"github.com/localpaas/localpaas/localpaas_app/config"
	"github.com/localpaas/localpaas/localpaas_app/infra/database"
	"github.com/localpaas/localpaas/localpaas_app/infra/gocronqueue"
	"github.com/localpaas/localpaas/localpaas_app/infra/logging"
	"github.com/localpaas/localpaas/localpaas_app/infra/rediscache"
	"github.com/localpaas/localpaas/localpaas_app/repository"
	"github.com/localpaas/localpaas/localpaas_app/repository/cacherepository"
	"github.com/localpaas/localpaas/localpaas_app/service/cronjobservice"
	"github.com/localpaas/localpaas/localpaas_app/service/settingservice"
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
	cronJobService            cronjobservice.Service
	taskService               taskservice.Service
	settingService            settingservice.Service

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
	cronJobService cronjobservice.Service,
	taskService taskservice.Service,
	settingService settingservice.Service,
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
		cronJobService:            cronJobService,
		taskService:               taskService,
		settingService:            settingService,
	}
}

func (q *taskQueue) Start() (err error) {
	// Initialize task queue worker if configured
	if q.isWorkerMode() {
		q.logger.Infof("starting task queue server...")
		q.server, err = gocronqueue.NewServer(&gocronqueue.Config{
			TaskMap:                 q.taskExecutorMap,
			RedisClient:             q.redisClient,
			Logger:                  q.logger,
			Concurrency:             q.config.Tasks.Queue.Concurrency,
			TaskCheckInterval:       q.config.Tasks.Queue.TaskCheckInterval,
			TaskCheckFunc:           q.findSchedulingTasks,
			TaskCreateInterval:      q.config.Tasks.Queue.TaskCreateInterval,
			TaskCreateFunc:          q.doCreateTasksForJobs,
			TaskCanScheduleFunc:     q.canScheduleTask,
			HealthcheckBaseInterval: q.config.Tasks.Healthcheck.BaseInterval,
			HealthcheckFunc:         q.doHealthcheck,
		})
		if err != nil {
			return apperrors.Wrap(err)
		}

		go func() {
			if err = q.server.Start(); err != nil {
				q.logger.Errorf("failed to start task queue server: %v", err)
			}
		}()
	}

	// Initialize task queue client if configured
	if q.isAppMode() {
		q.logger.Infof("starting task queue client...")
		q.client, err = gocronqueue.NewClient(q.redisClient, q.logger)
		if err != nil {
			return apperrors.Wrap(err)
		}
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

func (q *taskQueue) isAppMode() bool {
	return q.config.RunMode == config.RunModeApp || q.config.RunMode == config.RunModeAppAndWorker
}

func (q *taskQueue) isWorkerMode() bool {
	return q.config.RunMode == config.RunModeWorker || q.config.RunMode == config.RunModeAppAndWorker ||
		q.config.RunMode == config.RunModeUpdater
}
