package logginguc

import (
	"context"
	"testing"
	"time"

	"github.com/moby/moby/api/types/swarm"
	"github.com/moby/moby/client"
	"github.com/stretchr/testify/assert"

	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/infra/database"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/bunex"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/obi"
	"github.com/hivepaas/hivepaas/hivepaas_app/repository"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/loggingservice"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/settings"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/systemsettings/logginguc/loggingdto"
	"github.com/hivepaas/hivepaas/services/docker"
)

// loggingSetting answers the logging settings given.
type loggingSetting struct {
	repository.SettingRepo
	logging *entity.LoggingSettings
}

func (s *loggingSetting) GetSingle(context.Context, database.IDB, *entity.ObjectScope, base.SettingType, bool,
	...bunex.SelectQueryOption) (*entity.Setting, error) {
	setting := &entity.Setting{Type: base.SettingTypeLogging, UpdateVer: 3}
	setting.MustSetData(s.logging)
	return setting, nil
}

// oneNode is a swarm of one node, of 8 GB.
type oneNode struct{ docker.Manager }

func (oneNode) NodeList(context.Context, ...docker.NodeListOption) (*client.NodeListResult, error) {
	node := swarm.Node{ID: "n1"}
	node.Description.Hostname = "vps-1"
	node.Description.Resources.MemoryBytes = 8 << 30
	return &client.NodeListResult{Items: []swarm.Node{node}}, nil
}

// statusLogging answers one node's status, and keeps how far back it was
// asked for.
type statusLogging struct {
	loggingservice.Service
	since time.Duration
}

func (l *statusLogging) PerformanceStatus(_ context.Context, _ database.IDB, since time.Duration) (
	map[string]*loggingservice.PerformanceNodeStatus, error) {
	l.since = since
	return map[string]*loggingservice.PerformanceNodeStatus{"n1": {Status: obi.Status{Node: "n1",
		Preflight: obi.Preflight{OK: true, Kernel: "6.8.0"}}}}, nil
}

// The nodes' statuses are read over the time their agents write one: every 10
// minutes while the feature is off - so that the nodes can be chosen knowing
// which can run OBI - every minute while it is on.
func TestGetLoggingPerformanceReadsTheStatusesByTheSwitch(t *testing.T) {
	cfg := &entity.LoggingSettings{Enabled: true, Sources: entity.LoggingSources{Apps: true},
		Performance: &entity.LoggingPerformance{Nodes: []*entity.LoggingPerformanceNode{{ID: "n1"}}}}
	logs := &statusLogging{}
	uc := &UC{BaseUC: &settings.BaseUC{SettingRepo: &loggingSetting{logging: cfg}}, dockerManager: oneNode{},
		loggingService: logs}
	resp, err := uc.GetLoggingPerformance(context.Background(), nil, loggingdto.NewGetLoggingPerformanceReq())
	if !assert.NoError(t, err) {
		return
	}
	data := resp.Data
	assert.False(t, data.Enabled)
	assert.True(t, data.LogsStored)
	assert.Equal(t, 3, data.UpdateVer)
	assert.Empty(t, data.StatusReason)
	assert.Equal(t, 12*time.Minute, logs.since)
	if assert.Len(t, data.Nodes, 1) {
		assert.Equal(t, "vps-1", data.Nodes[0].Hostname)
		assert.True(t, data.Nodes[0].Enabled, "chosen, kept while off")
		assert.Equal(t, "medium", data.Nodes[0].Recommended)
		if assert.NotNil(t, data.Nodes[0].Status) {
			assert.True(t, data.Nodes[0].Status.OK)
		}
	}

	cfg.Performance.Enabled = true
	_, err = uc.GetLoggingPerformance(context.Background(), nil, loggingdto.NewGetLoggingPerformanceReq())
	assert.NoError(t, err)
	assert.Equal(t, 3*time.Minute, logs.since)

	// The logs not stored: nothing to read.
	cfg.Enabled, logs.since = false, 0
	resp, err = uc.GetLoggingPerformance(context.Background(), nil, loggingdto.NewGetLoggingPerformanceReq())
	assert.NoError(t, err)
	assert.Equal(t, loggingdto.PerformanceStatusReasonLogsNotStored, resp.Data.StatusReason)
	assert.Zero(t, logs.since)
}
