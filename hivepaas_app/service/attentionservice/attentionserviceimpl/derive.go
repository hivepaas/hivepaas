package attentionserviceimpl

import (
	"cmp"
	"slices"
	"strings"
	"time"

	"github.com/moby/moby/api/types/swarm"

	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/attentionservice"
	"github.com/hivepaas/hivepaas/services/docker"
)

const (
	// restartWindow and restartThreshold say what "keeps restarting" is: this
	// many failed tasks this recently. One failure is a crash; three in an hour
	// is a loop swarm will not get out of on its own.
	restartWindow    = time.Hour
	restartThreshold = 3

	// startGrace is how long a service that was just created or changed is left
	// to start before its missing tasks count: every deployment is, for a
	// moment, an app not running.
	startGrace = 2 * time.Minute
)

// clusterState is what the items are derived from, read once for all of them.
type clusterState struct {
	now      time.Time
	apps     []*entity.App
	services []swarm.Service
	tasks    []swarm.Task
	nodes    []swarm.Node
}

// deriveItems is what needs attention in a cluster in that state. It reads only
// what it is given, so what it concludes can be tested without a cluster.
func deriveItems(state *clusterState) []*attentionservice.Item {
	appsByService := make(map[string]*entity.App, len(state.apps))
	for _, app := range state.apps {
		if app.ServiceID != "" {
			appsByService[app.ServiceID] = app
		}
	}
	tasksByService := make(map[string][]swarm.Task, len(state.services))
	for _, task := range state.tasks {
		tasksByService[task.ServiceID] = append(tasksByService[task.ServiceID], task)
	}

	var items []*attentionservice.Item
	for i := range state.services {
		svc := &state.services[i]
		scope, subject, projectName, ok := serviceScope(svc, appsByService[svc.ID])
		if !ok {
			continue
		}
		item := serviceItem(state.now, svc, tasksByService[svc.ID])
		if item == nil {
			continue
		}
		item.Scope, item.Subject, item.ProjectName = scope, subject, projectName
		items = append(items, item)
	}
	items = append(items, nodeItems(state)...)

	slices.SortStableFunc(items, func(a, b *attentionservice.Item) int {
		return cmp.Or(
			cmp.Compare(severityRank(a.Severity), severityRank(b.Severity)),
			cmp.Compare(a.Subject, b.Subject),
		)
	})
	return items
}

// serviceScope is who a service's items are for. An app's are for whoever may
// read its env, and HivePaaS's own - its hidden project's apps, and the services
// of the stack it is installed as - for whoever may read the system screens. A
// service that is neither is somebody else's, and says nothing here.
func serviceScope(
	svc *swarm.Service,
	app *entity.App,
) (scope attentionservice.Scope, subject, projectName string, ok bool) {
	if app != nil {
		if app.Status != base.AppStatusActive || app.Project == nil || app.ProjectEnv == nil {
			return scope, "", "", false
		}
		if app.Project.Key == base.HivepaasProjectKey {
			return attentionservice.Scope{Type: attentionservice.ScopeSystem}, app.Name, "", true
		}
		return attentionservice.Scope{
			Type:       attentionservice.ScopeApp,
			ProjectID:  app.ProjectID,
			ProjectEnv: app.ProjectEnv.Key,
			AppID:      app.ID,
		}, app.Name, app.Project.Name, true
	}
	if svc.Spec.Labels[docker.StackLabelNamespace] == base.HivepaasProjectKey {
		name := strings.TrimPrefix(svc.Spec.Name, base.HivepaasProjectKey+"_")
		return attentionservice.Scope{Type: attentionservice.ScopeSystem}, name, "", true
	}
	return scope, "", "", false
}

// serviceItem is what is wrong with a service, if anything: tasks failing over
// and over, or fewer running than asked for. A loop is the one reported when
// both hold, since it is why the tasks are missing.
func serviceItem(now time.Time, svc *swarm.Service, tasks []swarm.Task) *attentionservice.Item {
	if svc.UpdateStatus != nil && (svc.UpdateStatus.State == swarm.UpdateStateUpdating ||
		svc.UpdateStatus.State == swarm.UpdateStateRollbackStarted) {
		return nil // being deployed: missing tasks are expected
	}

	var restarts int
	var latest *swarm.Task
	var firstRecent time.Time
	for i := range tasks {
		task := &tasks[i]
		if task.Status.State != swarm.TaskStateFailed && task.Status.State != swarm.TaskStateRejected {
			continue
		}
		if latest == nil || task.Status.Timestamp.After(latest.Status.Timestamp) {
			latest = task
		}
		if now.Sub(task.Status.Timestamp) <= restartWindow {
			restarts++
			if firstRecent.IsZero() || task.Status.Timestamp.Before(firstRecent) {
				firstRecent = task.Status.Timestamp
			}
		}
	}
	lastError := ""
	if latest != nil {
		lastError = cmp.Or(latest.Status.Err, latest.Status.Message)
	}

	if restarts >= restartThreshold {
		return &attentionservice.Item{
			Kind:      attentionservice.KindAppRestarting,
			Severity:  attentionservice.SeverityCritical,
			Restarts:  restarts,
			LastError: lastError,
			Since:     firstRecent,
		}
	}

	status := svc.ServiceStatus
	if status == nil || status.DesiredTasks == 0 || status.RunningTasks >= status.DesiredTasks {
		return nil
	}
	if now.Sub(svc.UpdatedAt) < startGrace {
		return nil
	}
	since := svc.UpdatedAt
	if latest != nil && latest.Status.Timestamp.After(since) {
		since = latest.Status.Timestamp
	}
	return &attentionservice.Item{
		Kind:      attentionservice.KindAppNotRunning,
		Severity:  attentionservice.SeverityCritical,
		Running:   status.RunningTasks,
		Desired:   status.DesiredTasks,
		LastError: lastError,
		Since:     since,
	}
}

// nodeItems is what is wrong with the nodes: one that is down, and one whose
// tasks are allowed, by their limits, more memory than it has. A task without a
// memory limit adds nothing to the sum, so the sum is the least they may use.
func nodeItems(state *clusterState) []*attentionservice.Item {
	limits := make(map[string]int64, len(state.nodes))
	for i := range state.tasks {
		task := &state.tasks[i]
		if task.NodeID == "" || task.DesiredState != swarm.TaskStateRunning {
			continue
		}
		if res := task.Spec.Resources; res != nil && res.Limits != nil {
			limits[task.NodeID] += res.Limits.MemoryBytes
		}
	}

	var items []*attentionservice.Item
	for i := range state.nodes {
		node := &state.nodes[i]
		scope := attentionservice.Scope{Type: attentionservice.ScopeCluster}
		hostname := node.Description.Hostname
		if node.Status.State != swarm.NodeStateReady {
			items = append(items, &attentionservice.Item{
				Kind:      attentionservice.KindNodeDown,
				Severity:  attentionservice.SeverityCritical,
				Scope:     scope,
				Subject:   hostname,
				NodeState: string(node.Status.State),
				Since:     node.UpdatedAt,
			})
			continue
		}
		total := node.Description.Resources.MemoryBytes
		if total > 0 && limits[node.ID] > total {
			items = append(items, &attentionservice.Item{
				Kind:         attentionservice.KindNodeOvercommitted,
				Severity:     attentionservice.SeverityWarning,
				Scope:        scope,
				Subject:      hostname,
				MemoryLimits: limits[node.ID],
				MemoryTotal:  total,
			})
		}
	}
	return items
}

func severityRank(severity attentionservice.Severity) int {
	switch severity {
	case attentionservice.SeverityCritical:
		return 0
	case attentionservice.SeverityWarning:
		return 1
	default:
		return 2 //nolint:mnd
	}
}
