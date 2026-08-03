package queueimpl

import (
	"context"

	"github.com/hivepaas/hivepaas/hivepaas_app/apperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/config"
	"github.com/hivepaas/hivepaas/hivepaas_app/infra/database"
	"github.com/hivepaas/hivepaas/hivepaas_app/infra/gocronqueue"
	"github.com/hivepaas/hivepaas/hivepaas_app/infra/logging"
	"github.com/hivepaas/hivepaas/hivepaas_app/infra/rediscache"
	"github.com/hivepaas/hivepaas/hivepaas_app/repository"
	"github.com/hivepaas/hivepaas/hivepaas_app/repository/cacherepository"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/schedjobservice"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/settingservice"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/startupservice"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/taskservice"
	"github.com/hivepaas/hivepaas/hivepaas_app/tasks/queue"
)

type taskQueue struct {
	db          *database.DB
	config      *config.Config
	logger      logging.Logger
	server      *gocronqueue.Server
	client      *gocronqueue.Client
	redisClient rediscache.Client

	settingRepo          repository.SettingRepo
	taskRepo             repository.TaskRepo
	taskInfoRepo         cacherepository.TaskInfoRepo
	periodicSettingsRepo cacherepository.PeriodicSettingsRepo

	schedJobService schedjobservice.Service
	settingService  settingservice.Service
	startupService  startupservice.Service
	taskService     taskservice.Service

	taskExecutorMap  map[base.TaskType]gocronqueue.TaskExecFunc
	rawExecutors     map[base.TaskType]queue.TaskExecFunc
	periodicExecutor queue.PeriodicExecFunc
}

func New(
	db *database.DB,
	config *config.Config,
	logger logging.Logger,
	redisClient rediscache.Client,
	settingRepo repository.SettingRepo,
	taskRepo repository.TaskRepo,
	cacheTaskInfoRepo cacherepository.TaskInfoRepo,
	periodicSettingsRepo cacherepository.PeriodicSettingsRepo,
	schedJobService schedjobservice.Service,
	taskService taskservice.Service,
	settingService settingservice.Service,
	startupService startupservice.Service,
) queue.TaskQueue {
	return &taskQueue{
		db:                   db,
		config:               config,
		logger:               logger,
		redisClient:          redisClient,
		settingRepo:          settingRepo,
		taskRepo:             taskRepo,
		taskInfoRepo:         cacheTaskInfoRepo,
		periodicSettingsRepo: periodicSettingsRepo,
		schedJobService:      schedJobService,
		taskService:          taskService,
		settingService:       settingService,
		startupService:       startupService,
	}
}

func (q *taskQueue) Start() (err error) {
	ctx := context.Background()
	lpSetting, err := q.startupService.LoadHivePaaSServiceSetting(ctx)
	if err != nil {
		return apperrors.Wrap(err)
	}
	lpSettings := lpSetting.MustAsHivePaaSService()

	runWorker := q.isWorkerMode()
	if q.isAppMode() {
		runWorker = lpSettings.WorkerSettings.RunWorkerInMainApp
	}

	// Initialize task queue worker if configured
	if runWorker {
		q.logger.Infof("starting task queue worker...")
		q.server, err = gocronqueue.NewServer(&gocronqueue.Config{
			TaskMap:              q.taskExecutorMap,
			RedisClient:          q.redisClient,
			Logger:               q.logger,
			Concurrency:          lpSettings.WorkerSettings.Concurrency,
			TaskCheckInterval:    lpSettings.TaskSettings.TaskCheckInterval.ToDuration(),
			TaskCheckFunc:        q.findSchedulingTasks,
			TaskCreateInterval:   lpSettings.TaskSettings.TaskCreateInterval.ToDuration(),
			TaskCreateFunc:       q.doCreateTasksForJobs,
			TaskCanScheduleFunc:  q.canScheduleTask,
			PeriodicBaseInterval: lpSettings.PeriodicSettings.BaseInterval.ToDuration(),
			PeriodicExecFunc:     q.doPeriodicJob,
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
		return apperrors.Wrap(apperrors.ErrUnavailable).WithParam("Name", "Task queue server")
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
		return apperrors.Wrap(apperrors.ErrUnavailable).WithParam("Name", "Task queue server")
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
