package dockerapiserviceimpl

import (
	"context"
	"errors"
	"testing"

	"github.com/moby/moby/api/types/swarm"
	"github.com/moby/moby/client"
	"github.com/stretchr/testify/assert"

	dockerapiclient "github.com/hivepaas/hivepaas/hivepaas_app/interface/agent/client/dockerapiservice"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/agentservice"
	"github.com/hivepaas/hivepaas/services/docker"
)

type fakeNodes struct {
	docker.Manager
	nodes []swarm.Node
}

func (f *fakeNodes) NodeList(_ context.Context, _ ...docker.NodeListOption) (*client.NodeListResult, error) {
	return &client.NodeListResult{Items: f.nodes}, nil
}

type fakeAgents struct {
	agentservice.Service
}

func (fakeAgents) GetAgentAddrForNode(_ context.Context, nodeID string) (string, error) {
	return "agent-on-" + nodeID, nil
}

// fakeAgentClient records what it is asked, and fails where told to.
type fakeAgentClient struct {
	addr  string
	calls *[]string
	fails bool
}

var errAgentDown = errors.New("agent down")

func (c *fakeAgentClient) Sync(context.Context) (int, error) {
	*c.calls = append(*c.calls, "sync "+c.addr)
	if c.fails {
		return 0, errAgentDown
	}
	return 1, nil
}

func (c *fakeAgentClient) RemoveApp(_ context.Context, appID string) (*dockerapiclient.RemoveAppResult, error) {
	*c.calls = append(*c.calls, "remove "+appID+" "+c.addr)
	return &dockerapiclient.RemoveAppResult{}, nil
}

func (c *fakeAgentClient) Close() error {
	return nil
}

func node(id string, state swarm.NodeState) swarm.Node {
	return swarm.Node{ID: id, Description: swarm.NodeDescription{Hostname: id}, Status: swarm.NodeStatus{State: state}}
}

func fanOutService(calls *[]string, failing string) *service {
	return &service{
		dockerManager: &fakeNodes{nodes: []swarm.Node{
			node("n1", swarm.NodeStateReady), node("n2", swarm.NodeStateReady), node("n3", swarm.NodeStateDown),
		}},
		agentService: fakeAgents{},
		agentClient: func(addr string) (dockerapiclient.DockerAPIServiceClient, error) {
			return &fakeAgentClient{addr: addr, calls: calls, fails: addr == "agent-on-"+failing}, nil
		},
	}
}

func TestSyncAgentsReachesEveryNodeThatIsUp(t *testing.T) {
	var calls []string
	assert.NoError(t, fanOutService(&calls, "").SyncAgents(context.Background()))
	assert.Equal(t, []string{"sync agent-on-n1", "sync agent-on-n2"}, calls)
}

func TestOneAgentFailingDoesNotKeepTheOthersFromHearing(t *testing.T) {
	var calls []string
	err := fanOutService(&calls, "n1").SyncAgents(context.Background())
	assert.ErrorIs(t, err, errAgentDown)
	assert.Contains(t, err.Error(), "node n1")
	assert.Equal(t, []string{"sync agent-on-n1", "sync agent-on-n2"}, calls)
}

func TestRemoveAppObjectsAsksEveryNode(t *testing.T) {
	var calls []string
	assert.NoError(t, fanOutService(&calls, "").RemoveAppObjects(context.Background(), "app1"))
	assert.Equal(t, []string{"remove app1 agent-on-n1", "remove app1 agent-on-n2"}, calls)
}
