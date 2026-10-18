package nodeuc

import (
	"context"
	"errors"
	"time"

	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/bunex"
)

// artifactsRunDelay is how long after a join the managed artifacts' job runs:
// time for the node to be Ready, and for its agent to start there.
const artifactsRunDelay = 2 * time.Minute

// scheduleArtifactsRun brings the managed artifacts' job forward after a node
// joined, rather than leave it to its next turn: an app on shared storage may
// start a task on the node, which has none of its artifacts yet. Nil when the
// job is off, or not made yet.
func (uc *UC) scheduleArtifactsRun(ctx context.Context, now time.Time) (*entity.Task, error) {
	job, err := uc.SettingRepo.GetSingle(ctx, uc.db, entity.NewObjectScopeGlobal(), base.SettingTypeSchedJob, true,
		bunex.SelectWhere("setting.kind = ?", base.SchedJobTypeManagedArtifacts))
	if errors.Is(err, hperrors.ErrNotFound) {
		return nil, nil
	}
	if err != nil {
		return nil, hperrors.Wrap(err)
	}
	task, err := uc.schedJobService.CreateSchedJobTask(job, now.Add(artifactsRunDelay), now)
	if err != nil {
		return nil, hperrors.Wrap(err)
	}
	if err = uc.taskRepo.Insert(ctx, uc.db, task); err != nil {
		return nil, hperrors.Wrap(err)
	}
	if err = uc.taskQueue.ScheduleTask(ctx, task); err != nil {
		return nil, hperrors.Wrap(err)
	}
	return task, nil
}
