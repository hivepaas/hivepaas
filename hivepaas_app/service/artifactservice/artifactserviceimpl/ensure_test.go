package artifactserviceimpl

import (
	"context"
	"testing"

	"github.com/moby/moby/api/types/swarm"
	"github.com/moby/moby/client"
	"github.com/stretchr/testify/assert"

	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
	agentartifactclient "github.com/hivepaas/hivepaas/hivepaas_app/interface/agent/client/artifactservice"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/agentservice"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/apptemplateservice/templatemodel"
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

func walg() *templatemodel.Artifact {
	return &templatemodel.Artifact{
		Name: "wal-g", Version: "3.0.8", MountPath: "/opt/hivepaas/bin/wal-g",
		Files: map[base.CPUArch]*templatemodel.ArtifactFile{
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

func TestEnsureOnNodePassesTheAgentsRefusalOn(t *testing.T) {
	svc, agent := newTestService("")
	agent.err = hperrors.Wrap(hperrors.ErrArtifactHashMismatch)

	_, err := svc.EnsureOnNode(context.Background(), &artifactservice.EnsureOnNodeReq{NodeID: "n1", Artifact: walg()})

	assert.ErrorIs(t, err, hperrors.ErrArtifactHashMismatch)
}
