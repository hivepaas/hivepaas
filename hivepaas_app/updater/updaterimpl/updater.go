package updaterimpl

import (
	"context"
	"os"
	"syscall"
	"time"

	"github.com/hivepaas/hivepaas/hivepaas_app/infra/database"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/logging"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/safego"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/hpappservice"
	"github.com/hivepaas/hivepaas/hivepaas_app/updater"
	"github.com/hivepaas/hivepaas/hivepaas_app/updater/tasksystemupdate"
)

// ensureAppRunningTimeout bounds the safety net below. It scales one service, so
// anything longer than this means the docker API is not answering and waiting
// changes nothing.
const ensureAppRunningTimeout = time.Minute

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
	safego.GoWithLogger(upd.logger, "updater.systemUpdate", func() {
		ctx := context.Background()
		_ = upd.systemUpdateExecutor.Execute(ctx, upd.db)
		// Before standing down, make sure something is left running. This covers
		// the update that ended without restoring the app - and also the run that
		// found no task at all, which is how a previous updater killed mid-update
		// shows up on the next boot.
		upd.ensureAppIsBack(ctx)
		// Shutdown the updater service (regardless of the update error)
		_ = upd.hpAppService.ShutdownHpUpdaterSwarmService(ctx)
		// Also send SIGTERM to the current process
		p, _ := os.FindProcess(os.Getpid())
		_ = p.Signal(syscall.SIGTERM)
	})
	return nil
}

// ensureAppIsBack is the last thing the updater does, and the only safety net
// there is for an update that stopped the main app and did not put it back.
//
// Failures are logged and swallowed: this runs after the update has already
// ended, and there is nothing further to abort.
func (upd *updaterImpl) ensureAppIsBack(ctx context.Context) {
	ctx, cancel := context.WithTimeout(ctx, ensureAppRunningTimeout)
	defer cancel()

	scaled, err := upd.hpAppService.EnsureHpAppRunning(ctx)
	if err != nil {
		upd.logger.Errorf("failed to check whether the main app is running: %v", err)
		return
	}
	if scaled {
		upd.logger.Warnf("the main app was left at zero replicas by an unfinished update, " +
			"scaled back to 1 - set the intended number from the dashboard")
	}
}

func (upd *updaterImpl) Shutdown() error {
	return nil
}
