package initializer

import (
	"github.com/localpaas/localpaas/localpaas_app/tasks/taskappdeploy"
	"github.com/localpaas/localpaas/localpaas_app/tasks/taskcronjobexec"
	"github.com/localpaas/localpaas/localpaas_app/tasks/taskdummy"
	"github.com/localpaas/localpaas/localpaas_app/tasks/taskhealthcheck"
	"github.com/localpaas/localpaas/localpaas_app/tasks/tasksystemupdate"
)

type WorkerInitializer struct {
}

// NOTE: these injections are required to make the task executor be available
func NewWorkerInitializer(
	_ *taskdummy.Executor,
	_ *taskappdeploy.Executor,
	_ *taskcronjobexec.Executor,
	_ *taskhealthcheck.Executor,
	_ *tasksystemupdate.Executor,
) *WorkerInitializer {
	return nil
}
