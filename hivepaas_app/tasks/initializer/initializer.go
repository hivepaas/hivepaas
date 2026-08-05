package initializer

import (
	"github.com/hivepaas/hivepaas/hivepaas_app/tasks/taskappclone"
	"github.com/hivepaas/hivepaas/hivepaas_app/tasks/taskappdeploy"
	"github.com/hivepaas/hivepaas/hivepaas_app/tasks/taskdummy"
	"github.com/hivepaas/hivepaas/hivepaas_app/tasks/taskperiodicjobexec"
	"github.com/hivepaas/hivepaas/hivepaas_app/tasks/taskschedjobexec"
	"github.com/hivepaas/hivepaas/hivepaas_app/tasks/taskworkflow"
)

type WorkerInitializer struct {
}

// NOTE: these injections are required to make the task executors be available
func NewWorkerInitializer(
	_ *taskdummy.Executor,
	_ *taskappdeploy.Executor,
	_ *taskappclone.Executor,
	_ *taskschedjobexec.Executor,
	_ *taskperiodicjobexec.Executor,
	_ *taskworkflow.Executor,
) *WorkerInitializer {
	return nil
}
