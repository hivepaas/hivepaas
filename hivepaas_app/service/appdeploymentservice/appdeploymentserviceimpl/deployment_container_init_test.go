package appdeploymentserviceimpl

import (
	"context"
	"testing"

	"github.com/moby/moby/api/types/swarm"
	"github.com/stretchr/testify/assert"

	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/tasklog"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/appdeploymentservice"
	"github.com/hivepaas/hivepaas/hivepaas_app/tasks/queue"
)

// An app with an entrypoint of its own is decided by that one, not by the
// image's: the image is not even read.
func TestContainerInitFollowsTheAppsOwnEntrypoint(t *testing.T) {
	data := &appDeploymentData{AppDeploymentReq: &appdeploymentservice.AppDeploymentReq{
		TaskExecData: &queue.TaskExecData{Task: &entity.Task{ID: "task-1"}, LogStore: tasklog.NewNullStore()},
	}}
	s := &service{} // no docker: reading the image would panic

	cs := &swarm.ContainerSpec{Image: "app:1", Command: []string{"tini", "--", "/app/run"}}
	s.applyContainerInit(context.Background(), data, cs)
	assert.False(t, *cs.Init, "an init of its own: not docker's too")

	cs = &swarm.ContainerSpec{Image: "app:1", Command: []string{"/app/run"}}
	s.applyContainerInit(context.Background(), data, cs)
	assert.True(t, *cs.Init, "the app itself: docker's, to reap what it orphans")
}
