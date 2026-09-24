package attentionserviceimpl

import (
	"testing"
	"time"

	"github.com/moby/moby/api/types/swarm"
	"github.com/stretchr/testify/assert"

	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/attentionservice"
	"github.com/hivepaas/hivepaas/services/docker"
)

var now = time.Date(2026, 9, 24, 12, 0, 0, 0, time.UTC)

func app(id, serviceID, projectKey string) *entity.App {
	return &entity.App{
		ID: id, Name: id, ServiceID: serviceID, ProjectID: "prj_" + projectKey, Status: base.AppStatusActive,
		Project:    &entity.Project{ID: "prj_" + projectKey, Key: projectKey, Name: projectKey},
		ProjectEnv: &entity.ProjectEnv{Key: "dev"},
	}
}

// svc is a service asking for desired tasks and running some, last changed long
// enough ago that missing ones count.
func svc(id string, running, desired uint64) swarm.Service {
	s := swarm.Service{ID: id, ServiceStatus: &swarm.ServiceStatus{RunningTasks: running, DesiredTasks: desired}}
	s.UpdatedAt = now.Add(-time.Hour)
	return s
}

func failedTask(serviceID string, ago time.Duration, err string) swarm.Task {
	return swarm.Task{ServiceID: serviceID, Status: swarm.TaskStatus{
		State: swarm.TaskStateFailed, Timestamp: now.Add(-ago), Err: err,
	}}
}

func kinds(items []*attentionservice.Item) []attentionservice.Kind {
	out := make([]attentionservice.Kind, 0, len(items))
	for _, item := range items {
		out = append(out, item.Kind)
	}
	return out
}

func TestAnAppWithMissingTasksIsNotRunning(t *testing.T) {
	items := deriveItems(&clusterState{
		now:      now,
		apps:     []*entity.App{app("a1", "svc1", "p2")},
		services: []swarm.Service{svc("svc1", 0, 1)},
		tasks:    []swarm.Task{failedTask("svc1", 5*time.Hour, "exit code 1")},
	})

	assert.Len(t, items, 1)
	item := items[0]
	assert.Equal(t, attentionservice.KindAppNotRunning, item.Kind)
	assert.Equal(t, attentionservice.Scope{
		Type: attentionservice.ScopeApp, ProjectID: "prj_p2", ProjectEnv: "dev", AppID: "a1",
	}, item.Scope)
	assert.Equal(t, uint64(0), item.Running)
	assert.Equal(t, uint64(1), item.Desired)
	assert.Equal(t, "exit code 1", item.LastError)
	assert.Equal(t, "p2", item.ProjectName)
}

// Some replicas running is still fewer than asked for, and another service's
// failures are not this one's.
func TestAnAppShortOfReplicasIsNotRunning(t *testing.T) {
	items := deriveItems(&clusterState{
		now:      now,
		apps:     []*entity.App{app("web", "svc1", "p1"), app("api", "svc2", "p1")},
		services: []swarm.Service{svc("svc1", 1, 2), svc("svc2", 1, 1)},
		tasks: []swarm.Task{
			failedTask("svc2", time.Minute, "api"), failedTask("svc2", 2*time.Minute, "api"),
			failedTask("svc2", 3*time.Minute, "api"),
		},
	})

	assert.Equal(t, []attentionservice.Kind{attentionservice.KindAppRestarting, attentionservice.KindAppNotRunning},
		kinds(items))
	assert.Equal(t, "api", items[0].Subject)
	assert.Equal(t, "web", items[1].Subject)
	assert.Equal(t, uint64(1), items[1].Running)
	assert.Empty(t, items[1].LastError, "the failures were api's")
}

// Three failures in the hour is a loop, and the loop is what is reported - it
// is why the tasks are missing.
func TestAnAppThatKeepsFailingIsRestarting(t *testing.T) {
	items := deriveItems(&clusterState{
		now:      now,
		apps:     []*entity.App{app("auth", "svc1", "p1")},
		services: []swarm.Service{svc("svc1", 0, 1)},
		tasks: []swarm.Task{
			failedTask("svc1", 50*time.Minute, "first"),
			failedTask("svc1", 20*time.Minute, "second"),
			failedTask("svc1", 2*time.Minute, "latest"),
			failedTask("svc1", 3*time.Hour, "too old to count"),
		},
	})

	assert.Equal(t, []attentionservice.Kind{attentionservice.KindAppRestarting}, kinds(items))
	assert.Equal(t, 3, items[0].Restarts)
	assert.Equal(t, "latest", items[0].LastError)
	assert.Equal(t, now.Add(-50*time.Minute), items[0].Since)
}

// What is in the middle of starting is not yet missing anything.
func TestAServiceBeingDeployedIsLeftToStart(t *testing.T) {
	updating := svc("svc1", 0, 1)
	updating.UpdateStatus = &swarm.UpdateStatus{State: swarm.UpdateStateUpdating}
	fresh := svc("svc2", 0, 1)
	fresh.UpdatedAt = now.Add(-30 * time.Second)

	items := deriveItems(&clusterState{
		now:      now,
		apps:     []*entity.App{app("a1", "svc1", "p1"), app("a2", "svc2", "p1")},
		services: []swarm.Service{updating, fresh},
	})

	assert.Empty(t, items)
}

// A stopped app asks for nothing, a disabled one is not looked at, and a
// service no app and no part of HivePaaS owns is somebody else's.
func TestWhatNobodyAskedToRunIsNotReported(t *testing.T) {
	disabled := app("a2", "svc2", "p1")
	disabled.Status = base.AppStatusDisabled

	items := deriveItems(&clusterState{
		now:      now,
		apps:     []*entity.App{app("a1", "svc1", "p1"), disabled},
		services: []swarm.Service{svc("svc1", 0, 0), svc("svc2", 0, 1), svc("svc3", 0, 1)},
	})

	assert.Empty(t, items)
}

// HivePaaS's own apps, and the services of the stack it is installed as, are
// the system screens' to show.
func TestHivePaaSOwnServicesAreForTheSystemScreens(t *testing.T) {
	traefik := svc("svc2", 0, 1)
	traefik.Spec.Name = "hivepaas_traefik"
	traefik.Spec.Labels = map[string]string{docker.StackLabelNamespace: "hivepaas"}

	items := deriveItems(&clusterState{
		now:      now,
		apps:     []*entity.App{app("registry", "svc1", base.HivepaasProjectKey)},
		services: []swarm.Service{svc("svc1", 0, 1), traefik},
	})

	assert.Len(t, items, 2)
	for _, item := range items {
		assert.Equal(t, attentionservice.Scope{Type: attentionservice.ScopeSystem}, item.Scope)
	}
	assert.ElementsMatch(t, []string{"registry", "traefik"}, []string{items[0].Subject, items[1].Subject})
}

func TestNodesDownAndOvercommitted(t *testing.T) {
	limited := func(nodeID string, bytes int64) swarm.Task {
		return swarm.Task{NodeID: nodeID, DesiredState: swarm.TaskStateRunning, Spec: swarm.TaskSpec{
			Resources: &swarm.ResourceRequirements{Limits: &swarm.Limit{MemoryBytes: bytes}},
		}}
	}
	node := func(id, hostname string, state swarm.NodeState, memory int64) swarm.Node {
		return swarm.Node{
			ID:          id,
			Description: swarm.NodeDescription{Hostname: hostname, Resources: swarm.Resources{MemoryBytes: memory}},
			Status:      swarm.NodeStatus{State: state},
		}
	}
	const gb = int64(1 << 30)

	items := deriveItems(&clusterState{
		now: now,
		tasks: []swarm.Task{
			limited("n1", 2*gb), limited("n1", gb),
			limited("n2", gb),
			{NodeID: "n1", DesiredState: swarm.TaskStateShutdown, Spec: limited("n1", 8*gb).Spec},
		},
		nodes: []swarm.Node{
			node("n1", "appstack-1", swarm.NodeStateReady, 2*gb),
			node("n2", "appstack-2", swarm.NodeStateReady, 4*gb),
			node("n3", "appstack-3", swarm.NodeStateDown, 4*gb),
		},
	})

	assert.Equal(t, []attentionservice.Kind{attentionservice.KindNodeDown, attentionservice.KindNodeOvercommitted},
		kinds(items), "critical first")
	assert.Equal(t, "appstack-3", items[0].Subject)
	assert.Equal(t, "appstack-1", items[1].Subject)
	assert.Equal(t, 3*gb, items[1].MemoryLimits, "a task being shut down no longer counts")
	assert.Equal(t, 2*gb, items[1].MemoryTotal)
	for _, item := range items {
		assert.Equal(t, attentionservice.ScopeCluster, item.Scope.Type)
	}
}
