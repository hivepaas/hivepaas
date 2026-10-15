package volumeserviceimpl

import (
	"context"
	"testing"

	"github.com/moby/moby/api/types/swarm"
	"github.com/moby/moby/client"
	"github.com/stretchr/testify/assert"

	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/agentservice"
	"github.com/hivepaas/hivepaas/services/docker"
)

// clusterOf is a cluster's nodes, by their labels and states.
type clusterOf struct {
	docker.Manager
	nodes []swarm.Node
}

func (c *clusterOf) NodeList(context.Context, ...docker.NodeListOption) (*client.NodeListResult, error) {
	return &client.NodeListResult{Items: c.nodes}, nil
}

func labeledNode(state swarm.NodeState, labels map[string]string) swarm.Node {
	return swarm.Node{Spec: swarm.NodeSpec{Annotations: swarm.Annotations{Labels: labels}},
		Status: swarm.NodeStatus{State: state}}
}

// askedAgents records the label an agent was asked for.
type askedAgents struct {
	agentservice.Service
	labels []string
}

func (a *askedAgents) GetAgentAddrForNodeLabel(_ context.Context, label string) (string, error) {
	a.labels = append(a.labels, label)
	return "", assert.AnError
}

func labelVolume(label string) *entity.Setting {
	setting := &entity.Setting{ID: "v1", RefID: "v1", Name: "data", Type: base.SettingTypeClusterVolume}
	setting.MustSetData(&entity.ClusterVolume{NodeLabel: label, Managed: true, Driver: "local"})
	return setting
}

// A volume pinned by a label no node of the cluster carries - a mistyped one -
// is on no node: removing it removes nothing, and asks no agent.
func TestAVolumeOfALabelNoNodeCarriesIsRemovedWithNothingToRemove(t *testing.T) {
	agents := &askedAgents{}
	svc := &service{dockerManager: &clusterOf{nodes: []swarm.Node{
		labeledNode(swarm.NodeStateReady, map[string]string{"disk": "ssd"}),
	}}, agentService: agents}

	err := svc.RemoveVolumeInCluster(context.Background(), labelVolume("disk=nvme"), false, true, 1, 0)

	assert.NoError(t, err)
	assert.Empty(t, agents.labels)
}

// A node carrying the label, though down, has the volume: its removal is
// asked for, and refused while the node cannot answer.
func TestAVolumeOfALabelADownNodeCarriesIsStillRemovedFromIt(t *testing.T) {
	agents := &askedAgents{}
	svc := &service{dockerManager: &clusterOf{nodes: []swarm.Node{
		labeledNode(swarm.NodeStateDown, map[string]string{"disk": "nvme"}),
	}}, agentService: agents}

	err := svc.RemoveVolumeInCluster(context.Background(), labelVolume("disk=nvme"), false, true, 1, 0)

	assert.Error(t, err)
	assert.Equal(t, []string{"disk=nvme"}, agents.labels)
}

func TestAnyNodeCarries(t *testing.T) {
	svc := &service{dockerManager: &clusterOf{nodes: []swarm.Node{
		labeledNode(swarm.NodeStateDown, map[string]string{"disk": "nvme", "gpu": ""}),
	}}}
	for _, tc := range []struct {
		label string
		want  bool
	}{{"disk=nvme", true}, {"disk", true}, {" disk = nvme ", true}, {"disk=ssd", false}, {"gpu", true},
		{"zone", false}} {
		label, want := tc.label, tc.want
		got, err := svc.anyNodeCarries(context.Background(), label)
		assert.NoError(t, err)
		assert.Equal(t, want, got, label)
	}
}
