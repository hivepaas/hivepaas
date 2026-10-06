package loggingserviceimpl

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	"github.com/moby/moby/api/types/swarm"
	"github.com/stretchr/testify/assert"

	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/obi"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/loggingservice"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/systemappservice"
	"github.com/hivepaas/hivepaas/services/logging"
)

// syncedService is the service with the logging apps running as cfg wants
// them, and cfg stored.
func syncedService(t *testing.T, fd *fakeDocker, cfg *entity.LoggingSettings) *service {
	t.Helper()
	s := newTestService(fd, storedSetting(t, cfg))
	_, err := s.Apply(context.Background(), nil, &loggingservice.SettingApplyReq{})
	assert.NoError(t, err)
	return s
}

func actionsOf(resp *loggingservice.SyncResp) map[string]entity.SystemAppSyncAction {
	out := map[string]entity.SystemAppSyncAction{}
	for _, app := range resp.Apps {
		out[app.Key] = app.Action
	}
	return out
}

// Apps as the settings want them are left alone.
func TestSyncLeavesAppsAsTheSettingsWantThem(t *testing.T) {
	s := syncedService(t, &fakeDocker{}, enabledConfig())
	appsOf(s).redeployed = nil

	resp, err := s.Sync(context.Background(), nil)

	assert.NoError(t, err)
	assert.Equal(t, map[string]entity.SystemAppSyncAction{backendAppKey: entity.SystemAppSyncNone,
		collectorAppKey: entity.SystemAppSyncNone}, actionsOf(resp))
	assert.Empty(t, resp.Tasks)
	assert.Empty(t, appsOf(s).redeployed)
}

// Logging switched off - by a save that did not confirm, or a row edited -
// takes both apps down, with their data: the stored logs stay.
func TestSyncRemovesAppsTheSettingsNoLongerWantKeepingTheirData(t *testing.T) {
	s := syncedService(t, &fakeDocker{}, enabledConfig())
	off := enabledConfig()
	off.Enabled = false
	settingsOf(s).setting = storedSetting(t, off)

	resp, err := s.Sync(context.Background(), nil)

	assert.NoError(t, err)
	assert.Equal(t, map[string]entity.SystemAppSyncAction{backendAppKey: entity.SystemAppSyncRemoved,
		collectorAppKey: entity.SystemAppSyncRemoved}, actionsOf(resp))
	assert.Equal(t, []string{collectorAppKey, backendAppKey}, appsOf(s).removed, "the collector first")
	assert.False(t, appsOf(s).removedStorage[backendAppKey], "the stored logs stay")
}

// With no settings at all, no app is wanted.
func TestSyncWithNoSettingsRemovesTheApps(t *testing.T) {
	s := syncedService(t, &fakeDocker{}, enabledConfig())
	settingsOf(s).setting = nil

	resp, err := s.Sync(context.Background(), nil)

	assert.NoError(t, err)
	assert.Equal(t, entity.SystemAppSyncRemoved, actionsOf(resp)[backendAppKey])
	assert.Equal(t, []string{collectorAppKey, backendAppKey}, appsOf(s).removed, "the collector first")
	assert.Empty(t, appsOf(s).apps)
	assert.False(t, appsOf(s).removedStorage[backendAppKey])
}

// Apps the settings want and are missing are provisioned, their deployments
// handed back to be scheduled.
func TestSyncProvisionsMissingApps(t *testing.T) {
	s := newTestService(&fakeDocker{}, storedSetting(t, enabledConfig()))

	resp, err := s.Sync(context.Background(), nil)

	assert.NoError(t, err)
	assert.Equal(t, map[string]entity.SystemAppSyncAction{backendAppKey: entity.SystemAppSyncProvisioned,
		collectorAppKey: entity.SystemAppSyncProvisioned}, actionsOf(resp))
	assert.Len(t, resp.Tasks, 2)
	assert.NotNil(t, resp.Cleanup)
}

// An app whose service is gone is removed and provisioned again - a deployment
// updates a service and cannot make one - its data kept.
func TestSyncRecreatesAnAppWhoseServiceIsGone(t *testing.T) {
	fd := &fakeDocker{}
	s := syncedService(t, fd, enabledConfig())
	delete(fd.inspected, appsOf(s).apps[backendAppKey].ServiceID)

	resp, err := s.Sync(context.Background(), nil)

	assert.NoError(t, err)
	assert.Equal(t, entity.SystemAppSyncRecreated, actionsOf(resp)[backendAppKey])
	assert.Equal(t, entity.SystemAppSyncNone, actionsOf(resp)[collectorAppKey])
	assert.Equal(t, []string{backendAppKey}, appsOf(s).removed)
	assert.False(t, appsOf(s).removedStorage[backendAppKey], "the stored logs stay")
	assert.Contains(t, fd.inspected, appsOf(s).apps[backendAppKey].ServiceID, "provisioned again")
}

// An app that drifted from the release is deployed with what the settings say.
func TestSyncDeploysAnAppThatDrifted(t *testing.T) {
	s := syncedService(t, &fakeDocker{}, enabledConfig())
	backend := appsOf(s).apps[backendAppKey]
	backend.Settings = []*entity.Setting{deploymentSetting("victoriametrics/victoria-logs:v1.40.0", "")}

	resp, err := s.Sync(context.Background(), nil)

	assert.NoError(t, err)
	assert.Equal(t, entity.SystemAppSyncUpdated, actionsOf(resp)[backendAppKey])
	assert.Len(t, resp.Tasks, 1)
}

// What a deployment would not mend is reported, and left as it is.
func TestSyncReportsWhatItLeavesToAPerson(t *testing.T) {
	s := syncedService(t, &fakeDocker{}, enabledConfig())
	appsOf(s).checks = map[string]*systemappservice.AppCheck{
		collectorAppKey: {Action: entity.SystemAppSyncReported, Problem: "its service is scaled to zero"},
	}

	resp, err := s.Sync(context.Background(), nil)

	assert.NoError(t, err)
	assert.Equal(t, entity.SystemAppSyncReported, actionsOf(resp)[collectorAppKey])
	assert.Empty(t, appsOf(s).removed)
}

// statusBackend answers the nodes' status rows.
type statusBackend struct {
	logging.Backend
	rows []*logging.OBIStatusRow
}

func (b *statusBackend) OBIStatus(_ context.Context, _ *logging.OBIStatusReq) ([]*logging.OBIStatusRow, error) {
	return b.rows, nil
}

func statusRow(t *testing.T, status obi.Status) *logging.OBIStatusRow {
	t.Helper()
	status.HP = obi.RowStatus
	return &logging.OBIStatusRow{Time: time.Now().UTC(), Msg: mustJSON(t, status)}
}

func node(id, hostname string) swarm.Node {
	return swarm.Node{ID: id, Description: swarm.NodeDescription{Hostname: hostname}}
}

func obiConfig(nodes ...string) *entity.LoggingSettings {
	cfg := enabledConfig()
	cfg.Performance = &entity.LoggingPerformance{Enabled: true}
	for _, id := range nodes {
		cfg.Performance.Nodes = append(cfg.Performance.Nodes, &entity.LoggingPerformanceNode{ID: id})
	}
	return cfg
}

func obiNodes(resp *loggingservice.SyncResp) map[string]*entity.OBINodeSyncOutput {
	out := map[string]*entity.OBINodeSyncOutput{}
	if resp.OBI != nil {
		for _, n := range resp.OBI.Nodes {
			out[n.NodeID] = n
		}
	}
	return out
}

// Each node the settings list is judged by its latest status; a node no longer
// in the cluster leaves them; one running OBI unasked has the agents read the
// settings again.
func TestSyncChecksOBIOnTheNodes(t *testing.T) {
	fd := &fakeDocker{
		nodes:      []swarm.Node{node("n1", "one"), node("n2", "two"), node("n3", "three"), node("n4", "four")},
		agentNodes: []string{"n1", "n2", "n4"},
	}
	s := syncedService(t, fd, obiConfig("n1", "n2", "n3", "gone"))
	b := &statusBackend{rows: []*logging.OBIStatusRow{
		statusRow(t, obi.Status{Node: "n1", Wanted: true, Running: true, Apps: 2, Preflight: obi.Preflight{OK: true}}),
		statusRow(t, obi.Status{Node: "n2", Wanted: true, Apps: 2,
			Preflight: obi.Preflight{Reasons: []string{obi.ReasonNoBTF}}}),
		statusRow(t, obi.Status{Node: "n4", Running: true, Apps: 2, Preflight: obi.Preflight{OK: true}}),
	}}
	s.newBackend = func(logging.BackendType, *logging.BackendConfig) (logging.Backend, error) { return b, nil }

	resp, err := s.Sync(context.Background(), nil)

	assert.NoError(t, err)
	nodes := obiNodes(resp)
	assert.Equal(t, entity.OBISyncNone, nodes["n1"].Action)
	assert.Equal(t, entity.OBISyncReported, nodes["n2"].Action)
	assert.Contains(t, nodes["n2"].Problem, obi.ReasonNoBTF)
	assert.Equal(t, entity.OBISyncReported, nodes["n3"].Action)
	assert.Equal(t, "its agent is not running", nodes["n3"].Problem)
	assert.Equal(t, entity.OBISyncCacheCleared, nodes["n4"].Action, "running, though not listed")
	assert.False(t, nodes["n4"].Expected)
	assert.Equal(t, entity.OBISyncNodeRemoved, nodes["gone"].Action)
	assert.True(t, resp.OBISettingsChanged)

	stored, err := settingsOf(s).setting.AsLoggingSettings()
	assert.NoError(t, err)
	ids := []string{}
	for _, n := range stored.Performance.Nodes {
		ids = append(ids, n.ID)
	}
	assert.Equal(t, []string{"n1", "n2", "n3"}, ids, "the node no longer in the cluster is taken off")
}

// A node's status read alone: what its agent says, and what is missing.
func TestJudgeOBINode(t *testing.T) {
	at := time.Now()
	status := func(s obi.Status) *loggingservice.PerformanceNodeStatus {
		return &loggingservice.PerformanceNodeStatus{Time: at, Status: s}
	}
	for _, tc := range []struct {
		name   string
		status *loggingservice.PerformanceNodeStatus
		agent  bool
		action entity.OBISyncAction
	}{
		{"running", status(obi.Status{Wanted: true, Running: true, Preflight: obi.Preflight{OK: true}}), true,
			entity.OBISyncNone},
		{"no app asks for it", status(obi.Status{Wanted: true, Preflight: obi.Preflight{OK: true}}), true,
			entity.OBISyncNone},
		{"the agent knows nothing of it", status(obi.Status{Preflight: obi.Preflight{OK: true}}), true,
			entity.OBISyncCacheCleared},
		{"it cannot run there", status(obi.Status{Wanted: true, Apps: 1}), true, entity.OBISyncReported},
		{"it does not run", status(obi.Status{Wanted: true, Apps: 1, Preflight: obi.Preflight{OK: true}}), true,
			entity.OBISyncReported},
		{"no status, the agent running", nil, true, entity.OBISyncReported},
		{"no status, no agent", nil, false, entity.OBISyncReported},
	} {
		t.Run(tc.name, func(t *testing.T) {
			assert.Equal(t, tc.action, judgeOBINode(tc.status, tc.agent).Action)
		})
	}
}

// OBI on with logging off cannot run: the settings are what is wrong.
func TestSyncSaysOBICannotRunWithTheLogsOff(t *testing.T) {
	cfg := obiConfig("n1")
	cfg.Enabled = false
	s := syncedService(t, &fakeDocker{nodes: []swarm.Node{node("n1", "one")}}, cfg)

	resp, err := s.Sync(context.Background(), nil)

	assert.NoError(t, err)
	if assert.NotNil(t, resp.OBI) {
		assert.Contains(t, resp.OBI.Problem, "OBI does not run")
		assert.Empty(t, resp.OBI.Nodes)
	}
}

func mustJSON(t *testing.T, v any) string {
	t.Helper()
	b, err := json.Marshal(v)
	if err != nil {
		t.Fatal(err)
	}
	return string(b)
}
