package internal

import (
	"context"
	"fmt"
	"time"

	"go.uber.org/fx"

	"github.com/hivepaas/hivepaas/hivepaas_app/config"
	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/infra/database"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/bunex"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/logging"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/timeutil"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/transaction"
	"github.com/hivepaas/hivepaas/hivepaas_app/repository"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/settinginitservice"
	"github.com/hivepaas/hivepaas/hivepaas_app/tasks/queue"
)

// DailyJobsTimezoneOnStart moves the daily system jobs to their times of day in
// the installation's timezone, when they were last given them in another - an
// installation from before there was one, in UTC, or one whose timezone has
// changed since. A job an administrator rescheduled is left as it is. The
// jobs moved are scheduled again, so it runs once the task queue has started.
// The app and the worker both start it; the system status, locked, has one of
// them move the jobs. A failure is logged: the next start tries again.
func DailyJobsTimezoneOnStart(
	lc fx.Lifecycle,
	cfg *config.Config,
	db *database.DB,
	sysStatusRepo repository.SystemStatusRepo,
	settingInitService settinginitservice.Service,
	taskQueue queue.TaskQueue,
	logger logging.Logger,
) {
	if cfg.RunMode == config.RunModeUpdater {
		return
	}
	lc.Append(fx.Hook{
		OnStart: func(ctx context.Context) error {
			err := moveDailyJobsToTimezone(ctx, db, sysStatusRepo, settingInitService, taskQueue,
				cfg.Location(), logger)
			if err != nil {
				logger.Errorf("failed to move the daily system jobs to the timezone %s: %v", cfg.Location(), err)
			}
			return nil
		},
	})
}

func moveDailyJobsToTimezone(
	ctx context.Context,
	db database.IDB,
	sysStatusRepo repository.SystemStatusRepo,
	settingInitService settinginitservice.Service,
	taskQueue queue.TaskQueue,
	to *time.Location,
	logger logging.Logger,
) error {
	err := transaction.Execute(ctx, db, func(tx database.Tx) error {
		sysStatus, err := sysStatusRepo.Get(ctx, tx, bunex.SelectFor("UPDATE"))
		if err != nil {
			return fmt.Errorf("failed to load the system status: %w", err)
		}
		from := time.UTC
		if sysStatus.ScheduleTimezone != "" {
			if from, err = time.LoadLocation(sysStatus.ScheduleTimezone); err != nil {
				return fmt.Errorf("the timezone the jobs were given their times in, %q: %w",
					sysStatus.ScheduleTimezone, err)
			}
		}
		if from.String() == to.String() {
			return nil
		}

		moved, left, err := settingInitService.MoveDailyJobs(ctx, tx, from, to)
		if err != nil {
			return fmt.Errorf("failed to move the daily jobs: %w", err)
		}
		if len(moved) > 0 {
			if err = taskQueue.ScheduleTasksForSchedJobs(ctx, tx, moved, true); err != nil {
				return fmt.Errorf("failed to schedule the daily jobs again: %w", err)
			}
		}

		sysStatus.ScheduleTimezone = to.String()
		sysStatus.UpdateVer++
		sysStatus.UpdatedAt = timeutil.NowUTC()
		err = sysStatusRepo.Upsert(ctx, tx, sysStatus,
			entity.SystemStatusUpsertingConflictCols, entity.SystemStatusScheduleTimezoneCols)
		if err != nil {
			return fmt.Errorf("failed to save the system status: %w", err)
		}
		logger.Info("moved the daily system jobs to the timezone", "from", from.String(), "to", to.String(),
			"moved", len(moved), "left as rescheduled", left)
		return nil
	})
	return hperrors.Wrap(err)
}
