package appdeploymentserviceimpl

import (
	"context"
	"fmt"

	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/tasklog"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/timeutil"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/schedjobtriggerservice"
)

const stepPreDeployJobs = "pre-deploy-jobs"

// deployStepPreDeployJobs runs the scheduled jobs listening to the app's
// pre-deploy, and waits for those a trigger holds the deploy for: one that
// fails, is canceled or runs out of time fails the deploy. Failing to fire fails
// it too: it cannot tell whether a job would have held it.
func (s *service) deployStepPreDeployJobs(ctx context.Context, data *appDeploymentData) (err error) {
	result, err := s.schedJobTriggerService.Fire(ctx, base.SchedJobTriggerPreDeploy, data.App,
		&schedjobtriggerservice.TriggerInfo{DeploymentID: data.Deployment.ID})
	if err != nil {
		return hperrors.Wrap(err)
	}
	if len(result.Runs) == 0 {
		return nil
	}

	data.Step = stepPreDeployJobs
	s.addStepStartLog(ctx, data, "Start running pre-deployment jobs...")
	defer s.addStepEndLog(ctx, data, timeutil.NowUTC(), err)
	for _, run := range result.Runs {
		if !run.Wait {
			_ = data.LogStore.Add(ctx, tasklog.NewOutFrame(fmt.Sprintf("Started job %s (run %s), not waited for",
				run.JobName, run.Task.ID), tasklog.TsNow))
		}
	}
	return hperrors.Wrap(s.schedJobTriggerService.WaitForRuns(ctx, result.WaitedRuns(), data.LogStore))
}

// fireDeployEndedEvent runs the jobs listening to how the deploy ended, once it
// is committed: post-deploy when it is done, deploy-failed when it failed, none
// when it was canceled. A failure is logged, never the deploy's.
func (s *service) fireDeployEndedEvent(ctx context.Context, data *appDeploymentData) {
	if data.Deployment == nil || data.App == nil {
		return
	}
	var event base.SchedJobTriggerEvent
	switch data.Deployment.Status { //nolint:exhaustive // the deploy has not ended otherwise
	case base.DeploymentStatusDone:
		event = base.SchedJobTriggerPostDeploy
	case base.DeploymentStatusFailed:
		event = base.SchedJobTriggerDeployFailed
	default:
		return
	}
	result, err := s.schedJobTriggerService.Fire(ctx, event, data.App,
		&schedjobtriggerservice.TriggerInfo{DeploymentID: data.Deployment.ID})
	if err != nil {
		_ = data.LogStore.Add(ctx, tasklog.NewWarnFrame(
			fmt.Sprintf("Failed to start the jobs listening to %s: %v", event, err), tasklog.TsNow))
		return
	}
	for _, run := range result.Runs {
		_ = data.LogStore.Add(ctx, tasklog.NewOutFrame(fmt.Sprintf("Started job %s (run %s) for %s",
			run.JobName, run.Task.ID, event), tasklog.TsNow))
	}
}
