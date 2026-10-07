package appdeploymentserviceimpl

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/tasklog"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/appdeploymentservice"
	"github.com/hivepaas/hivepaas/hivepaas_app/tasks/queue"
)

// A failed deployment says why, as a person reads it: an error of HivePaaS's
// own was its chain of codes alone - "ERR_PRECONDITION_FAILED ERR_MISSING" -
// with nothing of what was missing.
func TestAFailedDeploymentSaysWhy(t *testing.T) {
	deployment := &entity.Deployment{Output: &entity.AppDeploymentOutput{}}

	settleDeployment(deployment, false, hperrors.NewMissing("Registry auth to pull image"))

	assert.Equal(t, base.DeploymentStatusFailed, deployment.Status)
	assert.Equal(t, []string{"ERR_MISSING\nRegistry auth to pull image is missing"}, deployment.Output.Errors)
}

func TestAStepThatFailsLogsWhy(t *testing.T) {
	logs := tasklog.NewLocalStore("deployment")
	data := &appDeploymentData{AppDeploymentReq: &appdeploymentservice.AppDeploymentReq{
		TaskExecData: &queue.TaskExecData{Task: &entity.Task{ID: "task-1"}, LogStore: logs},
	}}

	(&service{}).addStepEndLog(context.Background(), data, time.Now(),
		hperrors.NewMissing("Registry auth to pull image"))

	frames, err := logs.GetData(context.Background(), 0)
	assert.NoError(t, err)
	if assert.Len(t, frames, 1) {
		assert.Contains(t, frames[0].Data, " with error: ERR_MISSING\nRegistry auth to pull image is missing")
	}
}
