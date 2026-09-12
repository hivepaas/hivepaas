package appplacementsettingsuc

import (
	"context"
	"testing"

	"github.com/moby/moby/api/types/swarm"
	"github.com/moby/moby/client"
	"github.com/stretchr/testify/assert"

	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/clusterservice"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/settings"
	"github.com/hivepaas/hivepaas/services/docker"
)

func node(role swarm.NodeRole, labels map[string]string) swarm.Node {
	return swarm.Node{Spec: swarm.NodeSpec{
		Role:        role,
		Annotations: swarm.Annotations{Labels: labels},
	}}
}

func TestHasMatchingNode(t *testing.T) {
	cluster := []swarm.Node{
		node(swarm.NodeRoleManager, map[string]string{"zone": "eu", "disk": "ssd"}),
		node(swarm.NodeRoleWorker, map[string]string{"zone": "us"}),
		node(swarm.NodeRoleWorker, map[string]string{"gpu": "true"}),
	}

	cases := []struct {
		name string
		cfg  entity.AppPlacementSettings
		want bool
	}{
		{"no rules", entity.AppPlacementSettings{}, true},
		{"a label some node carries", entity.AppPlacementSettings{
			RequireNodeLabels: []string{"zone=us"}}, true},
		{"a label nobody carries", entity.AppPlacementSettings{
			RequireNodeLabels: []string{"zone=apac"}}, false},
		// Swarm ANDs constraints: both must sit on one node, and here they do not.
		{"two labels on different nodes", entity.AppPlacementSettings{
			RequireNodeLabels: []string{"zone=us", "disk=ssd"}}, false},
		{"two labels on the same node", entity.AppPlacementSettings{
			RequireNodeLabels: []string{"zone=eu", "disk=ssd"}}, true},
		{"a bare key means true", entity.AppPlacementSettings{
			RequireNodeLabels: []string{"gpu"}}, true},
		{"excluding the only match", entity.AppPlacementSettings{
			RequireNodeLabels: []string{"zone=us"},
			ExcludeNodeLabels: []string{"zone=us"}}, false},
		// The manager carries the only disk=ssd label, so excluding managers
		// leaves the rule unsatisfiable.
		{"managers excluded", entity.AppPlacementSettings{
			ExcludeManagerNodes: true,
			RequireNodeLabels:   []string{"disk=ssd"}}, false},
		{"managers excluded, workers remain", entity.AppPlacementSettings{
			ExcludeManagerNodes: true}, true},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			assert.Equal(t, tc.want, hasMatchingNode(cluster, &tc.cfg))
		})
	}
}

func TestHasMatchingNodeOnAClusterOfManagersOnly(t *testing.T) {
	managers := []swarm.Node{node(swarm.NodeRoleManager, nil)}

	assert.False(t, hasMatchingNode(managers, &entity.AppPlacementSettings{ExcludeManagerNodes: true}))
}

type fakeClusterService struct {
	clusterservice.Service
	multiNode bool
}

func (f *fakeClusterService) IsMultiNode(_ context.Context) (bool, error) {
	return f.multiNode, nil
}

type fakeNodeLister struct {
	docker.Manager
	nodes  []swarm.Node
	called bool
}

func (f *fakeNodeLister) NodeList(_ context.Context, _ ...docker.NodeListOption) (*client.NodeListResult, error) {
	f.called = true
	return &client.NodeListResult{Items: f.nodes}, nil
}

func placementUC(multiNode bool, nodes []swarm.Node) (*UC, *fakeNodeLister) {
	lister := &fakeNodeLister{nodes: nodes}
	return &UC{
		dockerManager: lister,
		BaseUC:        &settings.BaseUC{ClusterService: &fakeClusterService{multiNode: multiNode}},
	}, lister
}

func TestCheckANodeMatchesRefusesRulesNothingSatisfies(t *testing.T) {
	uc, lister := placementUC(true, []swarm.Node{node(swarm.NodeRoleWorker, map[string]string{"zone": "eu"})})

	err := uc.checkANodeMatches(context.Background(),
		&entity.AppPlacementSettings{RequireNodeLabels: []string{"zone=apac"}})

	assert.ErrorIs(t, err, hperrors.ErrPlacementNoMatchingNode)
	assert.True(t, lister.called)
}

func TestCheckANodeMatchesAcceptsRulesSomeNodeSatisfies(t *testing.T) {
	uc, _ := placementUC(true, []swarm.Node{node(swarm.NodeRoleWorker, map[string]string{"zone": "eu"})})

	assert.NoError(t, uc.checkANodeMatches(context.Background(),
		&entity.AppPlacementSettings{RequireNodeLabels: []string{"zone=eu"}}))
}

// A single-node cluster never has these rules applied - loadPlacementSettingsData
// drops them, which is what the settings page tells the operator - so refusing to
// save them would refuse a rule that does nothing.
func TestCheckANodeMatchesSkipsASingleNodeCluster(t *testing.T) {
	uc, lister := placementUC(false, []swarm.Node{node(swarm.NodeRoleManager, nil)})

	err := uc.checkANodeMatches(context.Background(),
		&entity.AppPlacementSettings{RequireNodeLabels: []string{"zone=nowhere"}})

	assert.NoError(t, err)
	assert.False(t, lister.called, "no reason to ask docker about rules that are bypassed")
}

func TestCheckANodeMatchesIgnoresSettingsThatNarrowNothing(t *testing.T) {
	uc, lister := placementUC(true, nil)

	assert.NoError(t, uc.checkANodeMatches(context.Background(), &entity.AppPlacementSettings{}))
	assert.False(t, lister.called)
}
