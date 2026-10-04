package logginguc

import (
	"context"
	"testing"

	"github.com/moby/moby/api/types/swarm"
	"github.com/moby/moby/client"
	"github.com/stretchr/testify/assert"

	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/infra/database"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/bunex"
	"github.com/hivepaas/hivepaas/hivepaas_app/repository"
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

// While the feature is off, the nodes are listed for choosing, and nothing is
// asked of the logs: the agents write no status then. The logging service is
// nil here, and asking it would fail.
func TestGetLoggingPerformanceWhileOff(t *testing.T) {
	uc := &UC{BaseUC: &settings.BaseUC{SettingRepo: &loggingSetting{logging: &entity.LoggingSettings{
		Enabled: true, Sources: entity.LoggingSources{Apps: true},
		Performance: &entity.LoggingPerformance{Nodes: []*entity.LoggingPerformanceNode{{ID: "n1"}}},
	}}}, dockerManager: oneNode{}}
	resp, err := uc.GetLoggingPerformance(context.Background(), nil, loggingdto.NewGetLoggingPerformanceReq())
	if !assert.NoError(t, err) {
		return
	}
	data := resp.Data
	assert.False(t, data.Enabled)
	assert.True(t, data.LogsStored)
	assert.Equal(t, 3, data.UpdateVer)
	assert.Equal(t, loggingdto.PerformanceStatusReasonOff, data.StatusReason)
	if assert.Len(t, data.Nodes, 1) {
		assert.Equal(t, "vps-1", data.Nodes[0].Hostname)
		assert.True(t, data.Nodes[0].Enabled, "chosen, kept while off")
		assert.Equal(t, "medium", data.Nodes[0].Recommended)
		assert.Nil(t, data.Nodes[0].Status)
	}
}
