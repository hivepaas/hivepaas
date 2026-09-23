package specserviceimpl

import (
	"testing"

	"github.com/moby/moby/api/types/mount"
	"github.com/moby/moby/api/types/network"
	"github.com/moby/moby/api/types/swarm"
	"github.com/stretchr/testify/assert"

	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/unit"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/specservice/specmodel"
)

func blankTask() *swarm.TaskSpec {
	return &swarm.TaskSpec{ContainerSpec: &swarm.ContainerSpec{}}
}

// Every field export reads back out of a task's resources, built and read again.
func TestResourcesRoundTrip(t *testing.T) {
	swap, shm, swappiness := unit.DataSize(512<<20), unit.DataSize(64<<20), int64(10)
	want := &specmodel.Resources{
		Reservations: &specmodel.ResourceReservations{
			CPUs: 0.5, Memory: 256 << 20,
			GenericResources: []*specmodel.GenericResource{{Kind: "gpu", Value: "2"}, {Kind: "ssd", Value: "fast"}},
		},
		Limits: &specmodel.ResourceLimits{CPUs: 1, Memory: 512 << 20, Pids: 100},
		Memory: &specmodel.Memory{Swap: &swap, Swappiness: &swappiness, ShmSize: &shm},
		Capabilities: &specmodel.Capabilities{
			CapabilityAdd: []string{"NET_ADMIN"}, Sysctls: map[string]string{"vm.max_map_count": "262144"},
			Ulimits: []*specmodel.Ulimit{{Name: "nofile", Soft: 1024, Hard: 2048}}, OomScoreAdj: -100,
		},
	}
	task := blankTask()

	applyResources(want, task)

	assert.Equal(t, want, mapResources(task))
}

// A block the document leaves out clears what the task held, so building an
// exported document over a running service converges on the document.
func TestResourcesLeftOutAreCleared(t *testing.T) {
	task := blankTask()
	applyResources(&specmodel.Resources{
		Limits:       &specmodel.ResourceLimits{CPUs: 1},
		Memory:       &specmodel.Memory{ShmSize: new(unit.DataSize(64 << 20))},
		Capabilities: &specmodel.Capabilities{CapabilityAdd: []string{"NET_ADMIN"}},
	}, task)

	applyResources(nil, task)

	assert.Nil(t, mapResources(task))
	assert.Empty(t, task.ContainerSpec.Mounts, "the shared-memory mount goes with memory.shmSize")
}

func blankSpec() *swarm.ServiceSpec {
	return &swarm.ServiceSpec{TaskTemplate: *blankTask()}
}

// Attachments name their network; Docker takes a name where it takes an id, which
// is how the network an app is created on is named too.
func TestNetworksRoundTrip(t *testing.T) {
	want := &specmodel.Networks{
		Attachments:      []*specmodel.NetworkAttachment{{Name: "shop_prod", Aliases: []string{"api"}}},
		HostsFileEntries: []*specmodel.HostsFileEntry{{Address: "10.0.0.5", Hostnames: []string{"db", "db.local"}}},
		DNSConfig:        &specmodel.DNSConfig{Nameservers: []string{"1.1.1.1"}, Search: []string{"local"}},
		EndpointSpec: &specmodel.EndpointSpec{Mode: swarm.ResolutionModeVIP, Ports: []*specmodel.PortConfig{
			{Target: 53, Published: 53, Protocol: network.UDP, PublishMode: swarm.PortConfigPublishModeHost},
		}},
	}
	spec := blankSpec()

	assert.NoError(t, applyNetworks(want, spec))

	assert.Equal(t, want, mapNetworks(&spec.TaskTemplate, spec.EndpointSpec, nil))
}

// The network an app was created on stays when the document names none: a
// template never does, and every app has one.
func TestNetworksKeepTheAttachmentsWhenTheDocumentNamesNone(t *testing.T) {
	spec := blankSpec()
	spec.TaskTemplate.Networks = []swarm.NetworkAttachmentConfig{{Target: "shop_prod", Aliases: []string{"web"}}}
	spec.TaskTemplate.ContainerSpec.Hosts = []string{"10.0.0.5 db"}

	assert.NoError(t, applyNetworks(nil, spec))

	assert.Len(t, spec.TaskTemplate.Networks, 1)
	assert.Empty(t, spec.TaskTemplate.ContainerSpec.Hosts)
	assert.Nil(t, spec.EndpointSpec)
}

func TestNetworksRefuseANameserverThatIsNotAnAddress(t *testing.T) {
	dns := &specmodel.DNSConfig{Nameservers: []string{"dns.example"}}
	err := applyNetworks(&specmodel.Networks{DNSConfig: dns}, blankSpec())
	assert.Error(t, err)
}

// A Docker mount travels as Docker holds it, so building one is mapMount read
// backwards.
func TestDockerMountRoundTrip(t *testing.T) {
	for name, m := range map[string]mount.Mount{
		"a bind": {Type: mount.TypeBind, Source: "/etc/localtime", Target: "/etc/localtime", ReadOnly: true,
			BindOptions: &mount.BindOptions{Propagation: mount.PropagationRSlave, CreateMountpoint: true}},
		"a tmpfs": {Type: mount.TypeTmpfs, Target: "/cache",
			TmpfsOptions: &mount.TmpfsOptions{SizeBytes: 64 << 20, Mode: 0o1777}},
		"a volume mounted whole": {Type: mount.TypeVolume, Source: "shared", Target: "/shared",
			VolumeOptions: &mount.VolumeOptions{NoCopy: true, Labels: map[string]string{"a": "b"},
				DriverConfig: &mount.Driver{Name: "local", Options: map[string]string{"type": "nfs"}}}},
	} {
		t.Run(name, func(t *testing.T) {
			assert.Equal(t, m, toDockerMount(m.Target, mapMount(&m)))
		})
	}
}
