package systemappserviceimpl

import (
	"testing"

	"github.com/moby/moby/api/types/swarm"
	"github.com/stretchr/testify/assert"

	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
)

// A service that exists is judged by its task counts: short of them is
// reported, being updated is left for the next look.
func TestCheckService(t *testing.T) {
	replicated := swarm.ServiceMode{Replicated: &swarm.ReplicatedService{}}
	global := swarm.ServiceMode{Global: &swarm.GlobalService{}}
	for _, tc := range []struct {
		name    string
		svc     swarm.Service
		action  entity.SystemAppSyncAction
		problem string
	}{
		{"all running", swarm.Service{Spec: swarm.ServiceSpec{Mode: replicated},
			ServiceStatus: &swarm.ServiceStatus{RunningTasks: 1, DesiredTasks: 1}}, entity.SystemAppSyncNone, ""},
		{"scaled to zero", swarm.Service{Spec: swarm.ServiceSpec{Mode: replicated},
			ServiceStatus: &swarm.ServiceStatus{}}, entity.SystemAppSyncReported, "its service is scaled to zero"},
		{"no node for a global one", swarm.Service{Spec: swarm.ServiceSpec{Mode: global},
			ServiceStatus: &swarm.ServiceStatus{}}, entity.SystemAppSyncReported, "no node can run it"},
		{"tasks failing", swarm.Service{Spec: swarm.ServiceSpec{Mode: global},
			ServiceStatus: &swarm.ServiceStatus{RunningTasks: 1, DesiredTasks: 3}}, entity.SystemAppSyncReported,
			"1 of its 3 tasks are running"},
		{"being updated", swarm.Service{Spec: swarm.ServiceSpec{Mode: replicated},
			UpdateStatus:  &swarm.UpdateStatus{State: swarm.UpdateStateUpdating},
			ServiceStatus: &swarm.ServiceStatus{DesiredTasks: 1}}, entity.SystemAppSyncSkipped,
			"its service is being updated"},
		{"an update done", swarm.Service{Spec: swarm.ServiceSpec{Mode: replicated},
			UpdateStatus:  &swarm.UpdateStatus{State: swarm.UpdateStateCompleted},
			ServiceStatus: &swarm.ServiceStatus{RunningTasks: 1, DesiredTasks: 1}}, entity.SystemAppSyncNone, ""},
	} {
		t.Run(tc.name, func(t *testing.T) {
			got := checkService(&tc.svc)
			assert.Equal(t, tc.action, got.Action)
			assert.Equal(t, tc.problem, got.Problem)
			assert.False(t, got.ServiceGone)
		})
	}
}
