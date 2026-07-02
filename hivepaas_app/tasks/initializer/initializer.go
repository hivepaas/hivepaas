package initializer

import (
	"github.com/hivepaas/hivepaas/hivepaas_app/tasks/taskappdeploy"
	"github.com/hivepaas/hivepaas/hivepaas_app/tasks/taskdummy"
	"github.com/hivepaas/hivepaas/hivepaas_app/tasks/taskhealthcheck"
	"github.com/hivepaas/hivepaas/hivepaas_app/tasks/taskschedjobexec"
)

type WorkerInitializer struct {
}

// NOTE: these injections are required to make the task executors be available
func NewWorkerInitializer(
	_ *taskdummy.Executor,
	_ *taskappdeploy.Executor,
	_ *taskschedjobexec.Executor,
	_ *taskhealthcheck.Executor,
) *WorkerInitializer {
	return nil
}
