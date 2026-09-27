package schedjobtriggerserviceimpl

import (
	"context"
	"fmt"
	"time"

	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/config"
	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/bunex"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/tasklog"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/timeutil"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/schedjobtriggerservice"
)

const defaultWaitPollInterval = 3 * time.Second

func (s *service) WaitForRuns(
	ctx context.Context,
	runs []*schedjobtriggerservice.Run,
	logStore *tasklog.Store,
) error {
	if len(runs) == 0 {
		return nil
	}
	start := time.Now()
	pending := make(map[string]*schedjobtriggerservice.Run, len(runs))
	ids := make([]string, 0, len(runs))
	for _, run := range runs {
		pending[run.Task.ID] = run
		ids = append(ids, run.Task.ID)
		_ = logStore.Add(ctx, tasklog.NewOutFrame(fmt.Sprintf("Waiting for job %s (run %s)...",
			run.JobName, run.Task.ID), tasklog.TsNow))
	}

	for len(pending) > 0 {
		if err := timeutil.SleepCtx(ctx, pollInterval()); err != nil {
			s.cancelPending(ctx, pending)
			return hperrors.Wrap(err)
		}
		tasks, _, err := s.taskRepo.List(ctx, s.db, nil, nil, bunex.SelectWhere("task.id IN (?)", bunex.List(ids)))
		if err != nil {
			return hperrors.Wrap(err)
		}
		byID := make(map[string]*entity.Task, len(tasks))
		for _, task := range tasks {
			byID[task.ID] = task
		}
		for id, run := range pending {
			ended, err := runEnded(run, byID[id], time.Since(start))
			if err != nil {
				// The deploy's log says why it stops, in words: its error says it in codes.
				_ = logStore.Add(ctx, tasklog.NewErrFrame(hperrors.GetErrorDetail(err, ""), tasklog.TsNow))
				return err
			}
			if ended {
				_ = logStore.Add(ctx, tasklog.NewOutFrame(fmt.Sprintf("Job %s is done", run.JobName),
					tasklog.TsNow))
				delete(pending, id)
			}
		}
	}
	return nil
}

// runEnded says whether a waited run is done; an error when it will not be.
// A run failed with retries left is still going.
func runEnded(run *schedjobtriggerservice.Run, task *entity.Task, waited time.Duration) (bool, error) {
	switch {
	case task == nil:
		return false, hperrors.NewNotFound("Run of job " + run.JobName)
	case task.Status == base.TaskStatusDone:
		return true, nil
	case task.IsCanceled():
		return false, hperrors.NewArgumentInvalid("Pre-deploy job").
			WithExtraDetail("job %s was canceled", run.JobName)
	case task.IsFailedCompletely():
		return false, hperrors.NewArgumentInvalid("Pre-deploy job").
			WithExtraDetail("job %s failed", run.JobName)
	case run.Timeout > 0 && waited > run.Timeout:
		return false, hperrors.NewArgumentInvalid("Pre-deploy job").
			WithExtraDetail("job %s did not end within %s", run.JobName, run.Timeout)
	}
	return false, nil
}

// cancelPending cancels the runs a canceled deploy still waited for. Its own
// context has ended: the cancels get one that has not.
func (s *service) cancelPending(ctx context.Context, pending map[string]*schedjobtriggerservice.Run) {
	ctx = context.WithoutCancel(ctx)
	for id := range pending {
		_ = s.cancelRun(ctx, id)
	}
}

func pollInterval() time.Duration {
	if cfg := config.Current(); cfg != nil && cfg.Tasks.Triggers.WaitPollInterval > 0 {
		return cfg.Tasks.Triggers.WaitPollInterval
	}
	return defaultWaitPollInterval
}
