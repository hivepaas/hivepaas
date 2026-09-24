package dockerapiservice

import (
	"path"
	"testing"

	"github.com/moby/moby/api/types/mount"
	"github.com/moby/moby/api/types/swarm"
	"github.com/stretchr/testify/assert"

	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/dockerproxy"
)

func appSpec(mounts []mount.Mount, networks ...string) *swarm.ServiceSpec {
	spec := &swarm.ServiceSpec{TaskTemplate: swarm.TaskSpec{ContainerSpec: &swarm.ContainerSpec{Mounts: mounts}}}
	for _, network := range networks {
		spec.TaskTemplate.Networks = append(spec.TaskTemplate.Networks, swarm.NetworkAttachmentConfig{Target: network})
	}
	return spec
}

var dataMount = mount.Mount{Type: mount.TypeVolume, Source: "hp-vol-data", Target: "/data"}

func TestTheSocketIsMountedWhereTheProxyPutsIt(t *testing.T) {
	assert.Equal(t, path.Dir(dockerproxy.SocketPath), SocketMountTarget)
}

func TestAttachGivesTheSocketAndTheNetworkOnce(t *testing.T) {
	spec := appSpec([]mount.Mount{dataMount}, "project-net")
	Attach(spec, "app1", "net-app1")
	Attach(spec, "app1", "net-app1")

	assert.Equal(t, []mount.Mount{dataMount, SocketMount("app1")}, spec.TaskTemplate.ContainerSpec.Mounts)
	assert.Equal(t, []swarm.NetworkAttachmentConfig{{Target: "project-net"}, {Target: "net-app1"}},
		spec.TaskTemplate.Networks)
}

func TestAttachReplacesAnotherAppsSocket(t *testing.T) {
	spec := appSpec([]mount.Mount{dataMount, SocketMount("source-app")})
	Attach(spec, "copy", "net-copy")
	assert.Equal(t, []mount.Mount{dataMount, SocketMount("copy")}, spec.TaskTemplate.ContainerSpec.Mounts)
}

func TestDetachTakesTheSocketAndTheNetworkAway(t *testing.T) {
	spec := appSpec([]mount.Mount{dataMount, SocketMount("app1")}, "project-net", "net-app1")
	Detach(spec, "net-app1")
	assert.Equal(t, []mount.Mount{dataMount}, spec.TaskTemplate.ContainerSpec.Mounts)
	assert.Equal(t, []swarm.NetworkAttachmentConfig{{Target: "project-net"}}, spec.TaskTemplate.Networks)

	// An app that never had a network of its own keeps every network it has.
	spec = appSpec(nil, "project-net")
	Detach(spec, "")
	assert.Equal(t, []swarm.NetworkAttachmentConfig{{Target: "project-net"}}, spec.TaskTemplate.Networks)
}

func TestKeepSocketMountsCarriesTheSocketThroughAScreensSave(t *testing.T) {
	current := []mount.Mount{dataMount, SocketMount("app1")}
	other := mount.Mount{Type: mount.TypeVolume, Source: "hp-vol-data", Target: "/other"}
	// The screen does not show the socket, so what it saves lacks it.
	assert.Equal(t, []mount.Mount{other, SocketMount("app1")}, KeepSocketMounts([]mount.Mount{other}, current))
	// A client that sent it back anyway does not get it twice.
	assert.Equal(t, []mount.Mount{other, SocketMount("app1")},
		KeepSocketMounts([]mount.Mount{other, SocketMount("app1")}, current))
}

func TestKeepAppNetworkCarriesTheNetworkThroughAScreensSave(t *testing.T) {
	current := []swarm.NetworkAttachmentConfig{{Target: "project-net"}, {Target: "net-app1"}}
	final := []swarm.NetworkAttachmentConfig{{Target: "project-net", Aliases: []string{"web"}}}
	assert.Equal(t, []swarm.NetworkAttachmentConfig{{Target: "project-net", Aliases: []string{"web"}},
		{Target: "net-app1"}}, KeepAppNetwork(final, current, "net-app1"))
	assert.Equal(t, final, KeepAppNetwork(final, current, ""), "no network of its own, nothing to keep")
	assert.Equal(t, final, KeepAppNetwork(final, current[:1], "net-app1"), "not attached, not added")
}

func TestAppNetworkNamesAreRecognized(t *testing.T) {
	assert.True(t, IsAppNetworkName(NetworkName("app1")))
	assert.False(t, IsAppNetworkName("shop_prod_net"))
	socket := SocketMount("app1")
	assert.True(t, IsSocketMount(&socket))
	assert.False(t, IsSocketMount(&dataMount))
}

// Host mode binds the node's socket where apps look for it by default, and
// nothing of the proxy's.
func TestAttachHostGivesTheNodesSocketOnce(t *testing.T) {
	spec := appSpec([]mount.Mount{dataMount, SocketMount("app1")}, "project-net", "net-app1")
	AttachHost(spec)
	AttachHost(spec)

	assert.Equal(t, []mount.Mount{dataMount, HostSocketMount()}, spec.TaskTemplate.ContainerSpec.Mounts)
	assert.Equal(t, mount.Mount{Type: mount.TypeBind, Source: "/var/run/docker.sock", Target: "/var/run/docker.sock"},
		HostSocketMount())
}

func TestAttachAndDetachTakeTheNodesSocketAway(t *testing.T) {
	spec := appSpec([]mount.Mount{dataMount, HostSocketMount()})
	Attach(spec, "app1", "net-app1")
	assert.Equal(t, []mount.Mount{dataMount, SocketMount("app1")}, spec.TaskTemplate.ContainerSpec.Mounts)

	spec = appSpec([]mount.Mount{dataMount, HostSocketMount()})
	Detach(spec, "")
	assert.Equal(t, []mount.Mount{dataMount}, spec.TaskTemplate.ContainerSpec.Mounts)
}

// The storage screen shows neither socket and keeps both; a bind of the socket
// anywhere else is the app's own mount.
func TestTheNodesSocketIsASocketMountOnlyWhereHostModePutsIt(t *testing.T) {
	host := HostSocketMount()
	assert.True(t, IsSocketMount(&host))
	elsewhere := mount.Mount{Type: mount.TypeBind, Source: HostSocketPath, Target: "/docker.sock"}
	assert.False(t, IsSocketMount(&elsewhere))

	assert.Equal(t, []mount.Mount{dataMount, host}, KeepSocketMounts([]mount.Mount{dataMount}, []mount.Mount{host}))
}
