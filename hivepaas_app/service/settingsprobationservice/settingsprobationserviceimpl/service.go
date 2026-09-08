package settingsprobationserviceimpl

import (
	"time"

	"github.com/hivepaas/hivepaas/hivepaas_app/infra/database"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/logging"
	"github.com/hivepaas/hivepaas/hivepaas_app/repository"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/auditservice"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/settingsprobationservice"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/settingsrevertservice"
	"github.com/hivepaas/hivepaas/hivepaas_app/tasks/queue"
)

func New(
	db *database.DB,
	taskQueue queue.TaskQueue,
	logger logging.Logger,

	settingRepo repository.SettingRepo,
	taskRepo repository.TaskRepo,

	auditService auditservice.Service,
	settingsRevertService settingsrevertservice.Service,
) settingsprobationservice.Service {
	return &service{
		db:                    db,
		taskQueue:             taskQueue,
		logger:                logger,
		settingRepo:           settingRepo,
		taskRepo:              taskRepo,
		auditService:          auditService,
		settingsRevertService: settingsRevertService,
	}
}

// service holds no per-change state beyond the timers it arms.
type service struct {
	db        *database.DB
	taskQueue queue.TaskQueue
	logger    logging.Logger

	settingRepo repository.SettingRepo
	taskRepo    repository.TaskRepo

	auditService          auditservice.Service
	settingsRevertService settingsrevertservice.Service
}

const (
	// fallbackLag holds the in-process timer back so the queue, which is the path
	// with retries and a record, normally gets there first.
	fallbackLag = 20 * time.Second

	maxRetry    = 3
	retryDelay  = 15 * time.Second
	taskTimeout = 3 * time.Minute
)
