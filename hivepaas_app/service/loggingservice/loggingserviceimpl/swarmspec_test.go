package loggingserviceimpl

import (
	"testing"

	"github.com/moby/moby/api/types/mount"
	"github.com/stretchr/testify/assert"

	"github.com/hivepaas/hivepaas/services/logging"
)

func TestToSwarmServiceSpecGlobalModeLeavesReplicatedNil(t *testing.T) {
	rt := &logging.RuntimeSpec{Image: "img", Args: []string{"-a=1"}}

	spec, err := toSwarmServiceSpec(rt, swarmSpecOpts{Name: ServiceNameCollector, Global: true})
	if err != nil {
		t.Fatalf("toSwarmServiceSpec: %v", err)
	}

	assert.NotNil(t, spec.Mode.Global, "the collector must run one task per node")
	assert.Nil(t, spec.Mode.Replicated, "setting both modes is rejected by the daemon")
	assert.Equal(t, ServiceNameCollector, spec.Name)
	assert.Equal(t, "img", spec.TaskTemplate.ContainerSpec.Image)
	assert.Equal(t, []string{"-a=1"}, spec.TaskTemplate.ContainerSpec.Args)
}

func TestToSwarmServiceSpecReplicatedByDefault(t *testing.T) {
	rt := &logging.RuntimeSpec{Image: "img"}

	spec, err := toSwarmServiceSpec(rt, swarmSpecOpts{Name: ServiceNameBackend})
	if err != nil {
		t.Fatalf("toSwarmServiceSpec: %v", err)
	}

	if spec.Mode.Replicated == nil || spec.Mode.Replicated.Replicas == nil {
		t.Fatal("want one replica")
	}
	assert.Equal(t, uint64(1), *spec.Mode.Replicated.Replicas)
	assert.Nil(t, spec.Mode.Global)
}

// The backend's data is on one node's disk, so the constraint is required
// rather than a preference - the same shape the volume pinning emits.
func TestToSwarmServiceSpecPinsToANode(t *testing.T) {
	rt := &logging.RuntimeSpec{Image: "img"}

	spec, err := toSwarmServiceSpec(rt, swarmSpecOpts{Name: ServiceNameBackend, NodeID: "node-7"})
	if err != nil {
		t.Fatalf("toSwarmServiceSpec: %v", err)
	}

	if spec.TaskTemplate.Placement == nil {
		t.Fatal("no placement")
	}
	assert.Contains(t, spec.TaskTemplate.Placement.Constraints, "node.id==node-7")
}

func TestToSwarmServiceSpecTranslatesMounts(t *testing.T) {
	rt := &logging.RuntimeSpec{
		Image: "img",
		Mounts: []logging.Mount{
			{Source: "/var/lib/docker/containers", Target: "/var/lib/docker/containers", ReadOnly: true},
			{VolumeName: "vol-1", Target: "/data"},
		},
	}

	spec, err := toSwarmServiceSpec(rt, swarmSpecOpts{Name: "x"})
	if err != nil {
		t.Fatalf("toSwarmServiceSpec: %v", err)
	}

	mounts := spec.TaskTemplate.ContainerSpec.Mounts
	if len(mounts) != 2 {
		t.Fatalf("want two mounts, got %d", len(mounts))
	}
	assert.Equal(t, mount.TypeBind, mounts[0].Type)
	assert.True(t, mounts[0].ReadOnly)
	assert.Equal(t, mount.TypeVolume, mounts[1].Type)
	assert.Equal(t, "vol-1", mounts[1].Source)
}

// Everything this service creates carries the label, so a later version can
// find what it owns without guessing from names.
func TestToSwarmServiceSpecLabelsWhatItOwns(t *testing.T) {
	spec, err := toSwarmServiceSpec(&logging.RuntimeSpec{Image: "img"}, swarmSpecOpts{Name: "x"})
	if err != nil {
		t.Fatalf("toSwarmServiceSpec: %v", err)
	}

	assert.Equal(t, "true", spec.Labels[LabelManagedBy])
}

// The collector globs *-json.log. Anything this package runs writes `local`
// instead, so the stack cannot collect its own output - excluding it by path is
// impossible, because the path is a container id, not a service name.
func TestToSwarmServiceSpecKeepsTheStackOutOfTheCollector(t *testing.T) {
	spec, err := toSwarmServiceSpec(&logging.RuntimeSpec{Image: "img"}, swarmSpecOpts{Name: "x"})
	if err != nil {
		t.Fatalf("toSwarmServiceSpec: %v", err)
	}

	if spec.TaskTemplate.LogDriver == nil {
		t.Fatal("no log driver: the collector would read this service's own logs")
	}
	assert.Equal(t, logDriverLocal, spec.TaskTemplate.LogDriver.Name)
	assert.NotEqual(t, "json-file", spec.TaskTemplate.LogDriver.Name)
}

func TestToSwarmServiceSpecRequiresAnImage(t *testing.T) {
	_, err := toSwarmServiceSpec(&logging.RuntimeSpec{}, swarmSpecOpts{Name: "x"})

	assert.Error(t, err)
}
