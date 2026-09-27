package appdeploymentserviceimpl

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/tasklog"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/appdeploymentservice"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/schedjobtriggerservice"
	"github.com/hivepaas/hivepaas/hivepaas_app/tasks/queue"
)

type fakeTriggerService struct {
	schedjobtriggerservice.Service
	fired   []base.SchedJobTriggerEvent
	info    *schedjobtriggerservice.TriggerInfo
	result  *schedjobtriggerservice.FireResult
	waited  []*schedjobtriggerservice.Run
	waitErr error
}

func (f *fakeTriggerService) Fire(
	_ context.Context, event base.SchedJobTriggerEvent, _ *entity.App, info *schedjobtriggerservice.TriggerInfo,
) (*schedjobtriggerservice.FireResult, error) {
	f.fired = append(f.fired, event)
	f.info = info
	if f.result == nil {
		return &schedjobtriggerservice.FireResult{}, nil
	}
	return f.result, nil
}

func (f *fakeTriggerService) WaitForRuns(
	_ context.Context, runs []*schedjobtriggerservice.Run, _ *tasklog.Store,
) error {
	f.waited = runs
	return f.waitErr
}

func deployData(status base.DeploymentStatus) *appDeploymentData {
	return &appDeploymentData{
		AppDeploymentReq: &appdeploymentservice.AppDeploymentReq{
			TaskExecData: &queue.TaskExecData{LogStore: tasklog.NewNullStore()},
		},
		App:        &entity.App{ID: "a1"},
		Deployment: &entity.Deployment{ID: "d1", Status: status},
	}
}

// pre-deploy runs the jobs listening, and the deploy waits for those a trigger
// holds it for; one of them failing fails the deploy.
func TestPreDeployJobsHoldTheDeployForTheRunsThatWait(t *testing.T) {
	waited := &schedjobtriggerservice.Run{Task: &entity.Task{ID: "t1"}, JobName: "migrate", Wait: true}
	notWaited := &schedjobtriggerservice.Run{Task: &entity.Task{ID: "t2"}, JobName: "warm"}
	triggers := &fakeTriggerService{
		result:  &schedjobtriggerservice.FireResult{Runs: []*schedjobtriggerservice.Run{waited, notWaited}},
		waitErr: errors.New("job migrate failed"),
	}
	svc := &service{schedJobTriggerService: triggers}

	err := svc.deployStepPreDeployJobs(context.Background(), deployData(base.DeploymentStatusInProgress))

	assert.Error(t, err)
	assert.Equal(t, []base.SchedJobTriggerEvent{base.SchedJobTriggerPreDeploy}, triggers.fired)
	assert.Equal(t, "d1", triggers.info.DeploymentID)
	assert.Equal(t, []*schedjobtriggerservice.Run{waited}, triggers.waited)
}

func TestPreDeployJobsGoOnWithNoJobListening(t *testing.T) {
	triggers := &fakeTriggerService{}
	svc := &service{schedJobTriggerService: triggers}

	assert.NoError(t, svc.deployStepPreDeployJobs(context.Background(), deployData(base.DeploymentStatusInProgress)))
	assert.Nil(t, triggers.waited)
}

// A deploy that ended fires post-deploy or deploy-failed; a canceled one, nothing.
func TestFireDeployEndedEvent(t *testing.T) {
	cases := map[base.DeploymentStatus][]base.SchedJobTriggerEvent{
		base.DeploymentStatusDone:     {base.SchedJobTriggerPostDeploy},
		base.DeploymentStatusFailed:   {base.SchedJobTriggerDeployFailed},
		base.DeploymentStatusCanceled: nil,
	}
	for status, want := range cases {
		triggers := &fakeTriggerService{}
		svc := &service{schedJobTriggerService: triggers}

		svc.fireDeployEndedEvent(context.Background(), deployData(status))

		assert.Equal(t, want, triggers.fired, status)
	}
}
