package internal

import (
	"context"
	"fmt"
	"os"
	"strconv"
	"time"

	"go.uber.org/fx"

	"github.com/hivepaas/hivepaas/hivepaas_app/config"
	"github.com/hivepaas/hivepaas/hivepaas_app/infra/database"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/logging"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/safego"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/timeutil"
)

const (
	// WorkerHeartbeatFile is where the worker says it is still working. The path
	// is deliberately container-local rather than on the storage volume the app
	// and the worker share: a file on the volume outlives the container that
	// wrote it, so a dead worker would keep vouching for itself through the next
	// task's healthcheck.
	WorkerHeartbeatFile = "/tmp/hivepaas-worker.alive"

	workerHeartbeatInterval  = 15 * time.Second
	workerHeartbeatDBTimeout = 3 * time.Second
	workerHeartbeatFileMode  = 0o600
)

// InitWorkerHeartbeat gives the worker something a healthcheck can read.
//
// The worker serves no port, so there is nothing to probe from outside it. What
// it writes here is a unix timestamp, refreshed only while the process can still
// reach the database; the healthcheck in the compose file passes when that
// timestamp is recent. A worker whose process is up but whose database is gone -
// or whose goroutines are all blocked - stops refreshing it and is restarted.
//
// What it does NOT prove is that tasks are being executed. Nothing cheap does:
// the scheduler can be idle for hours quite legitimately. This catches the
// process being stuck, which is the failure swarm cannot see any other way,
// because restart_policy only ever notices a process that exits.
//
// NOTE: this must be invoked after InitTaskQueue. fx runs OnStart hooks in
// invoke order, and a heartbeat that starts before the queue does would report
// health for a worker that is not yet working.
func InitWorkerHeartbeat(
	lc fx.Lifecycle,
	cfg *config.Config,
	db *database.DB,
	logger logging.Logger,
) {
	if cfg.RunMode != config.RunModeWorker && cfg.RunMode != config.RunModeAppAndWorker {
		return
	}

	ctx, cancel := context.WithCancel(context.Background())
	lc.Append(fx.Hook{
		OnStart: func(_ context.Context) error {
			// Nothing is written here on purpose. The first beat lands one
			// interval in, so the file only ever exists once the process has
			// gone all the way through startup and reached the database.
			safego.Go("workerHeartbeat", func() {
				runWorkerHeartbeat(ctx, db, logger)
			})
			return nil
		},
		OnStop: func(_ context.Context) error {
			cancel()
			return nil
		},
	})
}

func runWorkerHeartbeat(ctx context.Context, db *database.DB, logger logging.Logger) {
	ticker := time.NewTicker(workerHeartbeatInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			// Leave the file behind rather than removing it. The container is
			// going away with it, and a shutdown racing the healthcheck should
			// not turn a clean stop into a failed probe.
			return
		case <-ticker.C:
			writeWorkerHeartbeat(ctx, db, logger)
		}
	}
}

// writeWorkerHeartbeat refreshes the file, or leaves it to go stale.
//
// A failure is logged and nothing else: the healthcheck is what acts on it, and
// it acts by counting how old the file is. Killing the process here instead
// would take away the very grace period the retries exist to give.
func writeWorkerHeartbeat(ctx context.Context, db *database.DB, logger logging.Logger) {
	pingCtx, cancel := context.WithTimeout(ctx, workerHeartbeatDBTimeout)
	defer cancel()

	if err := db.PingContext(pingCtx); err != nil {
		logger.Warnf("worker heartbeat skipped, database is not reachable: %v", err)
		return
	}

	stamp := strconv.FormatInt(timeutil.NowUTC().Unix(), 10)
	if err := writeFileAtomic(WorkerHeartbeatFile, stamp); err != nil {
		logger.Warnf("failed to write the worker heartbeat file %s: %v", WorkerHeartbeatFile, err)
	}
}

// writeFileAtomic replaces the file in one step, so a reader never catches it
// mid-write.
//
// WriteFile truncates first, leaving a window in which the healthcheck can read
// an empty file. The check reads empty as "never", so that window would count
// against the retries for no reason. Rename within one directory is atomic.
func writeFileAtomic(path, content string) error {
	tmpPath := path + ".tmp"
	if err := os.WriteFile(tmpPath, []byte(content), workerHeartbeatFileMode); err != nil {
		return fmt.Errorf("failed to write %s: %w", tmpPath, err)
	}
	if err := os.Rename(tmpPath, path); err != nil {
		_ = os.Remove(tmpPath)
		return fmt.Errorf("failed to move %s into place: %w", tmpPath, err)
	}
	return nil
}
