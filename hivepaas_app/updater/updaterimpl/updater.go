package updaterimpl

import (
	"context"
	"os"
	"syscall"

	"github.com/hivepaas/hivepaas/hivepaas_app/infra/database"
	"github.com/hivepaas/hivepaas/hivepaas_app/infra/logging"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/hpappservice"
	"github.com/hivepaas/hivepaas/hivepaas_app/updater"
	"github.com/hivepaas/hivepaas/hivepaas_app/updater/tasksystemupdate"
)

type updaterImpl struct {
	logger               logging.Logger
	db                   *database.DB
	hpAppService         hpappservice.Service
	systemUpdateExecutor *tasksystemupdate.Executor
}

func New(
	logger logging.Logger,
	db *database.DB,
	hpAppService hpappservice.Service,
	systemUpdateExecutor *tasksystemupdate.Executor,
) updater.Updater {
	e := &updaterImpl{
		logger:               logger,
		db:                   db,
		hpAppService:         hpAppService,
		systemUpdateExecutor: systemUpdateExecutor,
	}
	return e
}

func (upd *updaterImpl) Start() error {
	go func() {
		ctx := context.Background()
		_ = upd.systemUpdateExecutor.Execute(ctx, upd.db)
		// Shutdown the updater service (regardless of the update error)
		_ = upd.hpAppService.ShutdownHpUpdaterSwarmService(ctx)
		// Also send SIGTERM to the current process
		p, _ := os.FindProcess(os.Getpid())
		_ = p.Signal(syscall.SIGTERM)
	}()
	return nil
}

func (upd *updaterImpl) Shutdown() error {
	return nil
}
