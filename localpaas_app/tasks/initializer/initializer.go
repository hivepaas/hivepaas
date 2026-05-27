package initializer

import (
	"github.com/localpaas/localpaas/localpaas_app/tasks/taskappdeploy"
	"github.com/localpaas/localpaas/localpaas_app/tasks/taskdummy"
	"github.com/localpaas/localpaas/localpaas_app/tasks/taskhealthcheck"
	"github.com/localpaas/localpaas/localpaas_app/tasks/taskschedjobexec"
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
