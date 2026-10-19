package volumeserviceimpl

import (
	"context"
	"strings"
	"testing"
	"time"

	"github.com/moby/moby/api/types/mount"
	"github.com/moby/moby/api/types/swarm"
	"github.com/moby/moby/client"
	"github.com/stretchr/testify/assert"

	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/hpappservice"
	"github.com/hivepaas/hivepaas/services/docker"
)

// thisNode is n1's docker: it has no volume to inspect, runs helper containers
// here and helper tasks where Swarm places them, recording both.
type thisNode struct {
	docker.Manager
	containers int
	tasks      []*client.ServiceCreateOptions
}

func (n *thisNode) NodeCurrentID(context.Context) (string, error) { return "n1", nil }

func (n *thisNode) VolumeInspect(context.Context, string, ...docker.VolumeInspectOption) (
	*client.VolumeInspectResult, error) {
	return nil, hperrors.Wrap(hperrors.ErrNotFound)
}

func (n *thisNode) ContainerCreateToExec(context.Context, string, []string, ...docker.ContainerCreateOption) (
	*client.ContainerCreateResult, int64, error) {
	n.containers++
	return &client.ContainerCreateResult{}, 0, nil
}

func (n *thisNode) ServiceCreateToExec(_ context.Context, _ string, _ []string, _, _ time.Duration,
	options ...docker.ServiceCreateOption) (*client.ServiceCreateResult, int64, error) {
	// As the manager's does: the helper's container spec is there for options to fill.
	opts := &client.ServiceCreateOptions{}
	opts.Spec.TaskTemplate.ContainerSpec = &swarm.ContainerSpec{}
	for _, opt := range options {
		opt(opts)
	}
	n.tasks = append(n.tasks, opts)
	return &client.ServiceCreateResult{}, 0, nil
}

type agentImage struct{ hpappservice.Service }

func (agentImage) GetHpAgentImage(context.Context) string { return "hivepaas-agent:test" }

// A volume pinned to another node has an app's directory made there, by a task
// Swarm places by the pin; one pinned to this node, or to none, from here.
func TestAVolumesDirectoryIsMadeOnItsNode(t *testing.T) {
	volMount := &mount.Mount{Type: mount.TypeVolume, Source: "vol-1", Target: "/data",
		VolumeOptions: &mount.VolumeOptions{Subpath: "shop/prod/db"}}
	for name, tc := range map[string]struct {
		vol        *entity.ClusterVolume
		constraint string
	}{
		"pinned to another node": {&entity.ClusterVolume{NodeID: "n2"}, "node.id==n2"},
		"pinned by a label":      {&entity.ClusterVolume{NodeLabel: "zone=b"}, "node.labels.zone==b"},
		"pinned to this node":    {&entity.ClusterVolume{NodeID: "n1"}, ""},
		"every node's":           {&entity.ClusterVolume{}, ""},
	} {
		node := &thisNode{}
		svc := &service{dockerManager: node, hpAppService: agentImage{}}

		assert.NoError(t, svc.ensureVolumePermissionsOnNode(context.Background(), tc.vol, volMount, "shop/prod/db"),
			name)

		if tc.constraint == "" {
			assert.Equal(t, 1, node.containers, "%s: a helper container here", name)
			assert.Empty(t, node.tasks, name)
			continue
		}
		assert.Zero(t, node.containers, "%s: no helper here", name)
		if assert.Len(t, node.tasks, 1, name) {
			task := node.tasks[0].Spec.TaskTemplate
			assert.Equal(t, []string{tc.constraint}, task.Placement.Constraints, name)
			if assert.Len(t, task.ContainerSpec.Mounts, 1, name) {
				assert.Equal(t, "vol-1", task.ContainerSpec.Mounts[0].Source, name)
				assert.Equal(t, "/mnt/vol", task.ContainerSpec.Mounts[0].Target, name)
				assert.Empty(t, task.ContainerSpec.Mounts[0].VolumeOptions.Subpath, "%s: the volume's root", name)
			}
		}
	}
}

// A bind volume's directory on another node is made by that node's agent, on
// the node's filesystem as it sees it.
func TestABindVolumesDirectoryOnAnotherNodeIsMadeByItsAgent(t *testing.T) {
	exec := &fakeNodeExec{}
	svc := &service{dockerManager: &thisNode{}, nodeExecService: exec}

	err := svc.makeSubDirOnNode(context.Background(), &entity.ClusterVolume{NodeID: "n2"}, "/srv/data",
		"shop/prod/db", true)

	assert.NoError(t, err)
	if assert.NotNil(t, exec.req) {
		assert.Equal(t, "n2", exec.req.NodeID)
		script := strings.Join(exec.req.Command, " ")
		assert.Contains(t, script, "test -d /host/srv/data && ")
		assert.Contains(t, script, "mkdir -p /host/srv/data/shop/prod/db")
	}
}
