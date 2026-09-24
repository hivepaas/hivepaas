package dockerapiserviceimpl

import (
	"context"
	"errors"
	"fmt"

	"github.com/moby/moby/api/types/swarm"

	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
	dockerapiclient "github.com/hivepaas/hivepaas/hivepaas_app/interface/agent/client/dockerapiservice"
)

func (s *service) SyncAgents(ctx context.Context) error {
	return s.eachAgent(ctx, func(agent dockerapiclient.DockerAPIServiceClient) error {
		_, err := agent.Sync(ctx)
		return hperrors.Wrap(err)
	})
}

func (s *service) RemoveAppObjects(ctx context.Context, appID string) error {
	return s.eachAgent(ctx, func(agent dockerapiclient.DockerAPIServiceClient) error {
		_, err := agent.RemoveApp(ctx, appID)
		return hperrors.Wrap(err)
	})
}

// eachAgent calls fn with the agent of every node that is up. A node that
// cannot be reached, or whose agent fails, is reported, and does not keep the
// others from hearing.
func (s *service) eachAgent(ctx context.Context, fn func(dockerapiclient.DockerAPIServiceClient) error) error {
	nodes, err := s.dockerManager.NodeList(ctx)
	if err != nil {
		return hperrors.Wrap(err)
	}
	var errs []error
	for i := range nodes.Items {
		node := &nodes.Items[i]
		if node.Status.State != swarm.NodeStateReady {
			continue
		}
		if err = s.callAgent(ctx, node.ID, fn); err != nil {
			errs = append(errs, fmt.Errorf("node %s: %w", node.Description.Hostname, err))
		}
	}
	return errors.Join(errs...)
}

func (s *service) callAgent(ctx context.Context, nodeID string,
	fn func(dockerapiclient.DockerAPIServiceClient) error) error {
	addr, err := s.agentService.GetAgentAddrForNode(ctx, nodeID)
	if err != nil {
		return hperrors.Wrap(err)
	}
	agent, err := s.agentClient(addr)
	if err != nil {
		return hperrors.Wrap(err)
	}
	defer func() { _ = agent.Close() }()
	return fn(agent)
}
