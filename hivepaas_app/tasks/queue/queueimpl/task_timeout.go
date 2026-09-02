package queueimpl

import (
	"time"

	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
)

// taskTypeTimeouts caps how long a task of each type may run when it carries no timeout of its
// own. The cap matters beyond the task itself: a task holds its database transaction open for its
// whole run, and Postgres cannot reclaim any row version older than the longest open transaction.
// One task running for the old blanket three hours therefore bloats the entire database for three
// hours, not only the tasks table.
//
// The values are deliberately generous. Cutting a task short that used to finish is worse than
// the bloat it avoids, so types whose work is genuinely open-ended - cloning volumes, running
// whatever command a user configured - keep the long ceiling. What this buys is a much lower
// ceiling for the types whose duration is predictable, which are also the ones that run most.
const (
	timeoutDummy        = 5 * time.Minute
	timeoutPeriodicExec = 15 * time.Minute
	timeoutAppDeploy    = time.Hour
	timeoutSystemUpdate = time.Hour
)

var taskTypeTimeouts = map[base.TaskType]time.Duration{
	// A test task that should never be slow.
	base.TaskTypeDummy: timeoutDummy,
	// Health checks and similar probes. These usually carry an explicit timeout from the job
	// settings, so this only catches the ones that do not.
	base.TaskTypePeriodicExec: timeoutPeriodicExec,
	// Pull an image and wait for the swarm service to converge. Slow images make this minutes,
	// not hours; the image build itself waits on its own timeout.
	base.TaskTypeAppDeploy: timeoutAppDeploy,
	// Pull the new HivePaaS images and restart the services.
	base.TaskTypeSystemUpdate: timeoutSystemUpdate,

	// NOTE: the types below keep the long ceiling on purpose.
	//
	// App clone and preview rsync volume data, which is bounded by volume size rather than by
	// anything this code knows. Scheduled jobs run whatever command the user configured, and
	// workflows chain several steps that each have their own timeout. Lowering these would turn
	// a slow-but-working setup into a failing one.
	base.TaskTypeAppClone:     taskDefaultTimeout,
	base.TaskTypeAppPreview:   taskDefaultTimeout,
	base.TaskTypeSchedJobExec: taskDefaultTimeout,
	base.TaskTypeWorkflow:     taskDefaultTimeout,
}

// resolveTaskTimeout picks the timeout for a task: what the task asks for, else the ceiling for
// its type, else the blanket default.
func resolveTaskTimeout(task *entity.Task) time.Duration {
	if timeout := task.Config.Timeout.ToDuration(); timeout > 0 {
		return timeout
	}
	if timeout, ok := taskTypeTimeouts[task.Type]; ok && timeout > 0 {
		return timeout
	}
	return taskDefaultTimeout
}
