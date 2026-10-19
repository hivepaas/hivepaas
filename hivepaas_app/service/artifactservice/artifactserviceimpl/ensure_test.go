package artifactserviceimpl

import (
	"context"
	"testing"

	"github.com/moby/moby/api/types/swarm"
	"github.com/moby/moby/client"
	"github.com/stretchr/testify/assert"

	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
	agentartifactclient "github.com/hivepaas/hivepaas/hivepaas_app/interface/agent/client/artifactservice"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/agentservice"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/artifactservice"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecaseagent/artifactagentuc/artifactagentdto"
	"github.com/hivepaas/hivepaas/services/docker"
)

type fakeDocker struct {
	docker.Manager
	archByNode map[string]string
}

func (f *fakeDocker) NodeInspect(_ context.Context, nodeID string,
	_ ...docker.NodeInspectOption) (*client.NodeInspectResult, error) {
	arch, ok := f.archByNode[nodeID]
	if !ok {
		return nil, hperrors.Wrap(hperrors.ErrNotFound)
	}
	return &client.NodeInspectResult{Node: swarm.Node{ID: nodeID, Description: swarm.NodeDescription{
		Hostname: "node-" + nodeID, Platform: swarm.Platform{Architecture: arch, OS: "linux"},
	}}}, nil
}

type fakeAgents struct {
	agentservice.Service
}

func (fakeAgents) GetAgentAddrForNode(_ context.Context, nodeID string) (string, error) {
	return nodeID + ":10001", nil
}

// fakeAgentClient answers Ensure as an agent would, and records what it was asked.
type fakeAgentClient struct {
	addrs []string
	reqs  []*artifactagentdto.EnsureReq
	err   error
}

func (f *fakeAgentClient) dial(addr string) (agentartifactclient.ArtifactServiceClient, error) {
	f.addrs = append(f.addrs, addr)
	return f, nil
}

func (f *fakeAgentClient) Ensure(
	_ context.Context,
	req *artifactagentdto.EnsureReq,
) (*artifactagentdto.EnsureResp, error) {
	f.reqs = append(f.reqs, req)
	if f.err != nil {
		return nil, f.err
	}
	return &artifactagentdto.EnsureResp{HostPath: "/var/lib/hivepaas/cache/artifacts/sha256/" + req.SHA256,
		Downloaded: true}, nil
}

func (f *fakeAgentClient) Close() error { return nil }

func walg() *entity.AppArtifact {
	return &entity.AppArtifact{
		Name: "wal-g", Version: "3.0.8", MountPath: "/opt/hivepaas/bin/wal-g",
		Files: map[base.CPUArch]*entity.AppArtifactFile{
			base.CPUArchAMD64: {URL: "https://example.com/wal-g-amd64", SHA256: "aa", Size: 10},
			base.CPUArchARM64: {URL: "https://example.com/wal-g-aarch64", SHA256: "bb", Size: 20},
		},
	}
}

func newTestService(mirror string) (*service, *fakeAgentClient) {
	agent := &fakeAgentClient{}
	return &service{
		dockerManager: &fakeDocker{archByNode: map[string]string{"n1": "x86_64", "n2": "aarch64", "n3": "riscv64"}},
		agentService:  fakeAgents{},
		newClient:     agent.dial,
		mirror:        func() string { return mirror },
	}, agent
}

// The node's own architecture picks the file, and its agent is the one asked.
func TestEnsureOnNodePicksTheNodesArchitecture(t *testing.T) {
	for node, want := range map[string]struct {
		arch base.CPUArch
		url  string
	}{
		"n1": {base.CPUArchAMD64, "https://example.com/wal-g-amd64"},
		"n2": {base.CPUArchARM64, "https://example.com/wal-g-aarch64"},
	} {
		t.Run(node, func(t *testing.T) {
			svc, agent := newTestService("")

			resp, err := svc.EnsureOnNode(context.Background(),
				&artifactservice.EnsureOnNodeReq{NodeID: node, Artifact: walg()})

			assert.NoError(t, err)
			assert.Equal(t, want.arch, resp.Arch)
			assert.Equal(t, []string{node + ":10001"}, agent.addrs)
			assert.Equal(t, want.url, agent.reqs[0].URL)
			assert.Empty(t, agent.reqs[0].MirrorURL)
			assert.Equal(t, "/var/lib/hivepaas/cache/artifacts/sha256/"+agent.reqs[0].SHA256, resp.HostPath)
		})
	}
}

// A mirror holds files by their hash.
func TestEnsureOnNodeAsksTheMirrorByHash(t *testing.T) {
	svc, agent := newTestService("https://mirror.example.com/artifacts/")

	_, err := svc.EnsureOnNode(context.Background(), &artifactservice.EnsureOnNodeReq{NodeID: "n1", Artifact: walg()})

	assert.NoError(t, err)
	assert.Equal(t, "https://mirror.example.com/artifacts/aa", agent.reqs[0].MirrorURL)
}

func TestEnsureOnNodeRefusesANodeTheArtifactHasNoBuildFor(t *testing.T) {
	svc, agent := newTestService("")
	artifact := walg()
	delete(artifact.Files, base.CPUArchARM64)

	for node, arch := range map[string]string{"n2": "arm64", "n3": "riscv64"} {
		_, err := svc.EnsureOnNode(context.Background(), &artifactservice.EnsureOnNodeReq{NodeID: node, Artifact: artifact})

		assert.ErrorIs(t, err, hperrors.ErrArtifactArchUnsupported, node)
		info, _ := hperrors.ParseError(err, "en")
		assert.Contains(t, info.Detail, "node-"+node, "names the node")
		assert.Contains(t, info.Detail, arch)
	}
	assert.Empty(t, agent.reqs, "no agent is asked for a file that does not exist")
}

// A build asked for is had on any node, to be read there: the arm64 file on
// an amd64 node.
func TestEnsureOnNodeHasTheBuildAskedFor(t *testing.T) {
	svc, agent := newTestService("")

	resp, err := svc.EnsureOnNode(context.Background(), &artifactservice.EnsureOnNodeReq{NodeID: "n1",
		Artifact: walg(), Arch: base.CPUArchARM64})

	assert.NoError(t, err)
	assert.Equal(t, base.CPUArchARM64, resp.Arch)
	assert.Equal(t, "bb", agent.reqs[0].SHA256)
	assert.Equal(t, []string{"n1:10001"}, agent.addrs)
}

func TestEnsureOnNodePassesTheAgentsRefusalOn(t *testing.T) {
	svc, agent := newTestService("")
	agent.err = hperrors.Wrap(hperrors.ErrArtifactHashMismatch)

	_, err := svc.EnsureOnNode(context.Background(), &artifactservice.EnsureOnNodeReq{NodeID: "n1", Artifact: walg()})

	assert.ErrorIs(t, err, hperrors.ErrArtifactHashMismatch)
}

func (f *fakeDocker) NodeList(context.Context, ...docker.NodeListOption) (*client.NodeListResult, error) {
	var nodes []swarm.Node
	for _, id := range []string{"n1", "n2", "n3", "n4"} {
		arch := map[string]string{"n1": "x86_64", "n2": "aarch64", "n3": "riscv64", "n4": "x86_64"}[id]
		node := swarm.Node{ID: id,
			Spec: swarm.NodeSpec{Availability: swarm.NodeAvailabilityActive,
				Annotations: swarm.Annotations{Labels: map[string]string{"zone": id}}},
			Description: swarm.NodeDescription{Hostname: "node-" + id, Platform: swarm.Platform{Architecture: arch}},
			Status:      swarm.NodeStatus{State: swarm.NodeStateReady},
		}
		if id == "n4" {
			node.Status.State = swarm.NodeStateDown
		}
		nodes = append(nodes, node)
	}
	return &client.NodeListResult{Items: nodes}, nil
}

func specWith(constraints ...string) *swarm.ServiceSpec {
	return &swarm.ServiceSpec{TaskTemplate: swarm.TaskSpec{Placement: &swarm.Placement{Constraints: constraints}}}
}

func walgSettings() *entity.AppArtifactsSettings {
	return &entity.AppArtifactsSettings{Artifacts: []*entity.AppArtifact{walg()}}
}

// An app pinned to a node has its artifacts there, and only there.
func TestMountsForAPinnedApp(t *testing.T) {
	for _, constraint := range []string{"node.id==n2", "node.hostname==node-n2", "node.labels.zone==n2"} {
		svc, agent := newTestService("")

		mounts, err := svc.MountsFor(context.Background(), walgSettings(), specWith(constraint))

		assert.NoError(t, err, constraint)
		assert.Equal(t, []string{"n2:10001"}, agent.addrs, constraint)
		if assert.Len(t, mounts, 1) {
			assert.Equal(t, "/var/lib/hivepaas/cache/artifacts/sha256/bb", mounts[0].Source)
			assert.Equal(t, "/opt/hivepaas/bin/wal-g", mounts[0].Target)
			assert.True(t, mounts[0].ReadOnly)
		}
	}
}

// Without a pin, every node a task could start on needs the file, and one bind
// can only name one: nodes of two architectures are refused, with the reason.
func TestMountsForAnUnpinnedAppOnMixedNodesIsRefused(t *testing.T) {
	svc, _ := newTestService("")

	_, err := svc.MountsFor(context.Background(), walgSettings(), specWith("node.labels.zone!=n3"))

	assert.ErrorIs(t, err, hperrors.ErrArtifactArchUnsupported)
}

func TestMountsForAnUnpinnedAppOnLikeNodes(t *testing.T) {
	svc, agent := newTestService("")

	// n1 and n4 are both amd64, and n4 is down: only n1 can start a task.
	mounts, err := svc.MountsFor(context.Background(), walgSettings(), specWith("node.platform.arch==x86_64"))

	assert.NoError(t, err)
	assert.Equal(t, []string{"n1:10001"}, agent.addrs)
	assert.Len(t, mounts, 1)
}

func TestMountsForNoArtifactsAsksNothing(t *testing.T) {
	svc, agent := newTestService("")

	mounts, err := svc.MountsFor(context.Background(), nil, specWith("node.id==n1"))

	assert.NoError(t, err)
	assert.Empty(t, mounts)
	assert.Empty(t, agent.addrs)
}

func TestMountsForNoNodeIsRefused(t *testing.T) {
	svc, _ := newTestService("")

	_, err := svc.MountsFor(context.Background(), walgSettings(), specWith("node.id==gone"))

	assert.ErrorIs(t, err, hperrors.ErrArtifactDownloadFailed)
}
