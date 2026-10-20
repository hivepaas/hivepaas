package appservice

import (
	"testing"

	"github.com/moby/moby/api/types/swarm"
	"github.com/stretchr/testify/assert"
)

func specOn(networks ...swarm.NetworkAttachmentConfig) *swarm.ServiceSpec {
	return &swarm.ServiceSpec{TaskTemplate: swarm.TaskSpec{Networks: networks}}
}

// The alias goes on the network the app is reached by its key on, once.
func TestAddNetworkAlias(t *testing.T) {
	spec := specOn(swarm.NetworkAttachmentConfig{Target: "app-net"},
		swarm.NetworkAttachmentConfig{Target: "project-net", Aliases: []string{"shop-db"}})

	assert.True(t, AddNetworkAlias(spec, "shop-db", "shop-db-rw"))
	assert.Equal(t, []string{"shop-db", "shop-db-rw"}, spec.TaskTemplate.Networks[1].Aliases)
	assert.Empty(t, spec.TaskTemplate.Networks[0].Aliases)
	assert.False(t, AddNetworkAlias(spec, "shop-db", "shop-db-rw"), "it has it")

	// A service none of whose networks carries its key: its first network.
	spec = specOn(swarm.NetworkAttachmentConfig{Target: "project-net"})
	assert.True(t, AddNetworkAlias(spec, "shop-db-r1", "shop-db-rw"))
	assert.Equal(t, []string{"shop-db-rw"}, spec.TaskTemplate.Networks[0].Aliases)
	assert.False(t, AddNetworkAlias(specOn(), "shop-db", "shop-db-rw"), "no network to give it on")
}

func TestRemoveNetworkAlias(t *testing.T) {
	spec := specOn(swarm.NetworkAttachmentConfig{Target: "project-net", Aliases: []string{"shop-db", "shop-db-rw"}})

	assert.True(t, RemoveNetworkAlias(spec, "shop-db-rw"))
	assert.Equal(t, []string{"shop-db"}, spec.TaskTemplate.Networks[0].Aliases)
	assert.False(t, RemoveNetworkAlias(spec, "shop-db-rw"))
}
