package appdeploymentserviceimpl

import (
	"testing"

	"github.com/moby/moby/api/types/swarm"
	"github.com/stretchr/testify/assert"
)

// A build tags its image after the commit, so building the same commit again -
// with other build settings - gives the tag the service already runs: swarm
// finds nothing changed in the spec and keeps the old containers, unless the
// update is forced.
func TestApplyBuiltImage(t *testing.T) {
	spec := func(image string) *swarm.ServiceSpec {
		return &swarm.ServiceSpec{TaskTemplate: swarm.TaskSpec{ContainerSpec: &swarm.ContainerSpec{Image: image}}}
	}

	rebuilt := spec("shop-web:dev-36d81dd")
	applyBuiltImage(rebuilt, "shop-web:dev-36d81dd")
	assert.Equal(t, "shop-web:dev-36d81dd", rebuilt.TaskTemplate.ContainerSpec.Image)
	assert.Equal(t, uint64(1), rebuilt.TaskTemplate.ForceUpdate, "the same tag, built again")

	newCommit := spec("shop-web:dev-36d81dd")
	applyBuiltImage(newCommit, "shop-web:dev-a6a0fd1")
	assert.Equal(t, "shop-web:dev-a6a0fd1", newCommit.TaskTemplate.ContainerSpec.Image)
	assert.Equal(t, uint64(0), newCommit.TaskTemplate.ForceUpdate, "a tag of its own changes the spec")
}
