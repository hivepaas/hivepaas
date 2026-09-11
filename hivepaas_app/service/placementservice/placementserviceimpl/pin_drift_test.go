package placementserviceimpl

import (
	"context"
	"errors"
	"fmt"
	"testing"

	"github.com/moby/moby/api/types/swarm"
	"github.com/moby/moby/client"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/logging"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/placementservice"
	"github.com/hivepaas/hivepaas/services/docker"
)

func runningTask(nodeID string) swarm.Task {
	return swarm.Task{NodeID: nodeID, Status: swarm.TaskStatus{State: swarm.TaskStateRunning}}
}

// Applying the pin moves this task. Saying so beats an app that quietly restarts
// somewhere else and finds a directory that is not the one it was using.
func TestDriftingNodesReportsATaskElsewhere(t *testing.T) {
	assert.Equal(t, []string{"node-2"},
		driftingNodes([]swarm.Task{runningTask("node-2")}, "node-1"))
}

func TestDriftingNodesIgnoresTasksAlreadyInPlace(t *testing.T) {
	assert.Empty(t, driftingNodes([]swarm.Task{runningTask("node-1")}, "node-1"))
}

// A task that is not running is not being moved off anything.
func TestDriftingNodesIgnoresTasksNotRunning(t *testing.T) {
	shutdown := swarm.Task{NodeID: "node-2", Status: swarm.TaskStatus{State: swarm.TaskStateShutdown}}

	assert.Empty(t, driftingNodes([]swarm.Task{shutdown}, "node-1"))
}

// Nothing to compare against: an unpinned volume constrains nothing, and a label
// pin names a set of nodes that a node id alone cannot be checked against.
func TestDriftingNodesSaysNothingWithoutANodeID(t *testing.T) {
	assert.Empty(t, driftingNodes([]swarm.Task{runningTask("node-2")}, ""))
}

func TestDriftingNodesReportsEachNodeOnce(t *testing.T) {
	tasks := []swarm.Task{runningTask("node-2"), runningTask("node-2"), runningTask("node-3")}

	assert.ElementsMatch(t, []string{"node-2", "node-3"}, driftingNodes(tasks, "node-1"))
}

// warningLogger keeps the warnings so a test can assert one was emitted. The
// embedded interface leaves every other level nil, which is the point: this
// path is only ever supposed to warn.
type warningLogger struct {
	logging.Logger
	warnings []string
}

func (l *warningLogger) Warnf(template string, args ...any) {
	l.warnings = append(l.warnings, fmt.Sprintf(template, args...))
}

// taskListingDockerManager embeds the interface so only the one method this
// path reaches has a body; anything else panics rather than quietly passing.
type taskListingDockerManager struct {
	docker.Manager
	tasks  []swarm.Task
	err    error
	called int
}

func (m *taskListingDockerManager) ServiceTaskList(
	_ context.Context, _ string, _ []swarm.TaskState, _ ...docker.TaskListOption,
) (*client.TaskListResult, error) {
	m.called++
	if m.err != nil {
		return nil, m.err
	}
	return &client.TaskListResult{Items: m.tasks}, nil
}

func driftData(serviceID string, pins []placementservice.VolumePin) *placementSettingsData {
	return &placementSettingsData{
		ApplyPlacementSettingsReq: &placementservice.ApplyPlacementSettingsReq{
			Service: &swarm.Service{
				ID:   serviceID,
				Spec: swarm.ServiceSpec{Annotations: swarm.Annotations{Name: "shop-web"}},
			},
			VolumePins: pins,
		},
	}
}

func pinnedToNode1() []placementservice.VolumePin {
	return []placementservice.VolumePin{{VolumeName: "pgdata", NodeID: "node-1"}}
}

// The operator learns about the move before it happens, from the place that
// decided it - not from an app that came back somewhere else.
func TestWarnOnPinDriftNamesTheAppTheNodeAndThePin(t *testing.T) {
	dockerMgr := &taskListingDockerManager{tasks: []swarm.Task{runningTask("node-2")}}
	logger := &warningLogger{}
	svc := &service{dockerManager: dockerMgr, logger: logger}

	svc.warnOnPinDrift(context.Background(), driftData("svc-1", pinnedToNode1()))

	require.Len(t, logger.warnings, 1)
	assert.Contains(t, logger.warnings[0], "shop-web")
	assert.Contains(t, logger.warnings[0], "node-2")
	assert.Contains(t, logger.warnings[0], "node-1")
}

func TestWarnOnPinDriftIsQuietWhenTheTaskIsAlreadyOnThePinnedNode(t *testing.T) {
	dockerMgr := &taskListingDockerManager{tasks: []swarm.Task{runningTask("node-1")}}
	logger := &warningLogger{}
	svc := &service{dockerManager: dockerMgr, logger: logger}

	svc.warnOnPinDrift(context.Background(), driftData("svc-1", pinnedToNode1()))

	assert.Empty(t, logger.warnings)
}

// Conflicting pins are refused rather than applied, so nothing is about to be
// moved and there is nothing to warn about.
func TestWarnOnPinDriftIsQuietWhenThePinsConflict(t *testing.T) {
	dockerMgr := &taskListingDockerManager{tasks: []swarm.Task{runningTask("node-3")}}
	logger := &warningLogger{}
	svc := &service{dockerManager: dockerMgr, logger: logger}

	svc.warnOnPinDrift(context.Background(), driftData("svc-1", []placementservice.VolumePin{
		{VolumeName: "pgdata", NodeID: "node-1"},
		{VolumeName: "uploads", NodeID: "node-2"},
	}))

	assert.Empty(t, logger.warnings)
	assert.Zero(t, dockerMgr.called, "a conflict is settled before any task is listed")
}

// A label pin names a set of nodes. Asking docker which nodes carry the label,
// only to say a task may move, would spend calls on a warning that cannot be
// substantiated from a task list alone.
func TestWarnOnPinDriftDoesNotListTasksForALabelPin(t *testing.T) {
	dockerMgr := &taskListingDockerManager{tasks: []swarm.Task{runningTask("node-2")}}
	logger := &warningLogger{}
	svc := &service{dockerManager: dockerMgr, logger: logger}

	svc.warnOnPinDrift(context.Background(), driftData("svc-1",
		[]placementservice.VolumePin{{VolumeName: "pgdata", NodeLabel: "storage=fast"}}))

	assert.Empty(t, logger.warnings)
	assert.Zero(t, dockerMgr.called)
}

// An app being created has no service yet, so it has no task anywhere to move.
func TestWarnOnPinDriftDoesNotListTasksWithoutAServiceID(t *testing.T) {
	dockerMgr := &taskListingDockerManager{}
	logger := &warningLogger{}
	svc := &service{dockerManager: dockerMgr, logger: logger}

	svc.warnOnPinDrift(context.Background(), driftData("", pinnedToNode1()))

	assert.Empty(t, logger.warnings)
	assert.Zero(t, dockerMgr.called)
}

// The warning is a courtesy on the way to applying the settings; docker being
// unreachable must not turn it into a failure.
func TestWarnOnPinDriftStaysSilentWhenTasksCannotBeListed(t *testing.T) {
	dockerMgr := &taskListingDockerManager{err: errors.New("docker is away")}
	logger := &warningLogger{}
	svc := &service{dockerManager: dockerMgr, logger: logger}

	svc.warnOnPinDrift(context.Background(), driftData("svc-1", pinnedToNode1()))

	assert.Empty(t, logger.warnings)
}
