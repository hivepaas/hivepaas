package sysupdateserviceimpl

import (
	"context"
	"testing"

	"github.com/moby/moby/api/types/swarm"
	"github.com/stretchr/testify/assert"

	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/traefikservice"
)

func (f *fakeHpApp) GetHpAgentSwarmService(_ context.Context) (*swarm.Service, error) {
	if f.agent == nil {
		return nil, hperrors.Wrap(hperrors.ErrNotFound)
	}
	return f.agent, nil
}

// The cache is not what these tests are about: not deployed.
func (f *fakeHpApp) GetHpCacheSwarmService(_ context.Context) (*swarm.Service, error) {
	return nil, nil
}

// globalServiceOn is a service with a task on every node, as the agent is.
func globalServiceOn(id, image string) *swarm.Service {
	return &swarm.Service{
		ID: id,
		Spec: swarm.ServiceSpec{
			Annotations:  swarm.Annotations{Name: id},
			TaskTemplate: swarm.TaskSpec{ContainerSpec: &swarm.ContainerSpec{Image: image}},
			Mode:         swarm.ServiceMode{Global: &swarm.GlobalService{}},
		},
	}
}

// The agent is what the app asks to build, back up and restore on every node.
// An update that moves the app and leaves the agent behind runs a new app
// against an old agent.
func TestAgentUpdateMovesTheAgentWithRollbackArmed(t *testing.T) {
	f := &fakeDocker{}
	hp := &fakeHpApp{agent: globalServiceOn("hivepaas_agent", "hivepaas/hivepaas-agent:1.0.0-beta1")}
	s := &service{dockerManager: f, hpAppService: hp}
	data := loggingUpdateData(t, &base.ReleaseInfo{AgentImage: "hivepaas/hivepaas-agent:1.0.0-beta2"})

	assert.NoError(t, s.updateAgentService(context.Background(), data, argsOf(t, data)))

	assert.Equal(t, "hivepaas/hivepaas-agent:1.0.0-beta2", f.updated["hivepaas_agent"])
	spec := f.specs["hivepaas_agent"]
	if assert.NotNil(t, spec) && assert.NotNil(t, spec.UpdateConfig) {
		assert.Equal(t, swarm.UpdateFailureActionRollback, spec.UpdateConfig.FailureAction)
	}
	assert.NotNil(t, spec.Mode.Global, "the agent stays on every node")
}

// A release file from before agentImage existed names no agent: the agent is
// left as it is, as any component the release names no image for.
func TestAgentUpdateLeavesTheAgentOfAReleaseNamingNone(t *testing.T) {
	f := &fakeDocker{}
	hp := &fakeHpApp{agent: globalServiceOn("hivepaas_agent", "hivepaas/hivepaas-agent:1.0.0-beta1")}
	s := &service{dockerManager: f, hpAppService: hp}
	data := loggingUpdateData(t, &base.ReleaseInfo{AppImage: "hivepaas/hivepaas:1.0.0-beta2"})

	assert.NoError(t, s.updateAgentService(context.Background(), data, argsOf(t, data)))

	assert.Empty(t, f.updated)
}

// An installation without an agent service is not deployed, not broken.
func TestAgentUpdateSkipsAnInstallationWithoutOne(t *testing.T) {
	f := &fakeDocker{}
	s := &service{dockerManager: f, hpAppService: &fakeHpApp{}}
	data := loggingUpdateData(t, &base.ReleaseInfo{AgentImage: "hivepaas/hivepaas-agent:1.0.0-beta2"})

	assert.NoError(t, s.updateAgentService(context.Background(), data, argsOf(t, data)))

	assert.Empty(t, f.updated)
}

type fakeTraefik struct {
	traefikservice.Service
}

func (fakeTraefik) GetTraefikSwarmService(_ context.Context) (*swarm.Service, error) {
	return nil, nil
}

// The plan the dashboard shows before an update lists the agent with the rest.
func TestPlanListsTheAgent(t *testing.T) {
	hp := &fakeHpApp{
		app:   swarmServiceOn("hivepaas_app", "hivepaas/hivepaas:1.0.0-beta1", 1),
		agent: globalServiceOn("hivepaas_agent", "hivepaas/hivepaas-agent:1.0.0-beta1"),
	}
	s := &service{dockerManager: &fakeDocker{}, hpAppService: hp, traefikService: fakeTraefik{},
		systemAppService: &fakeSystemApps{apps: map[string]*entity.App{}}}

	plan, err := s.PlanUpdate(context.Background(), nil, &base.ReleaseInfo{
		AppImage: "hivepaas/hivepaas:1.0.0-beta2", AgentImage: "hivepaas/hivepaas-agent:1.0.0-beta2",
	})

	assert.NoError(t, err)
	var agent *struct{ current, target string }
	for _, c := range plan.Components {
		if c.Key == base.HivepaasAgentKey {
			agent = &struct{ current, target string }{c.CurrentImage, c.TargetImage}
		}
	}
	if assert.NotNil(t, agent, "the agent is in the plan") {
		assert.Equal(t, "hivepaas/hivepaas-agent:1.0.0-beta1", agent.current)
		assert.Equal(t, "hivepaas/hivepaas-agent:1.0.0-beta2", agent.target)
	}
}
