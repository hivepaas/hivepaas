package schedjobexecserviceimpl

import (
	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
)

// triggerEnv is what a run a trigger fired is told of its cause: the event, the
// app it happened to, and the deployment of a deploy event. Nothing for a run
// no trigger fired.
func triggerEnv(task *entity.Task) []string {
	if task == nil {
		return nil
	}
	args, err := task.ArgsAsSchedJobExec()
	if err != nil || args == nil || args.Trigger == nil {
		return nil
	}
	return []string{
		"HIVEPAAS_TRIGGER_EVENT=" + string(args.Trigger.Event),
		"HIVEPAAS_TRIGGER_APP=" + args.Trigger.AppID,
		"HIVEPAAS_TRIGGER_DEPLOYMENT=" + args.Trigger.DeploymentID,
	}
}
