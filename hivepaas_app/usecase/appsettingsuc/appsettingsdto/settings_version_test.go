package appsettingsdto

import (
	"testing"
	"time"

	"github.com/moby/moby/api/types/mount"
	"github.com/moby/moby/api/types/swarm"
	"github.com/stretchr/testify/assert"

	"github.com/hivepaas/hivepaas/hivepaas_app/service/dockerapiservice"
)

func serviceForVersions() *swarm.Service {
	replicas := uint64(2)
	return &swarm.Service{
		Meta: swarm.Meta{Version: swarm.Version{Index: 100}},
		Spec: swarm.ServiceSpec{
			Annotations: swarm.Annotations{Labels: map[string]string{"team": "shop"}},
			Mode:        swarm.ServiceMode{Replicated: &swarm.ReplicatedService{Replicas: &replicas}},
			TaskTemplate: swarm.TaskSpec{
				ContainerSpec: &swarm.ContainerSpec{
					Image:  "shop:1",
					Hosts:  []string{"10.0.0.1 db"},
					Mounts: []mount.Mount{{Type: mount.TypeVolume, Source: "data", Target: "/data"}},
				},
				Resources: &swarm.ResourceRequirements{Limits: &swarm.Limit{MemoryBytes: 64 << 20}},
				Placement: &swarm.Placement{Constraints: []string{"node.role==worker"}},
			},
		},
	}
}

type screenVersion struct {
	name    string
	version func(*swarm.Service) int
}

var screens = []screenVersion{
	{"resources", ResourceSettingsVersion},
	{"container", ContainerSettingsVersion},
	{"availability", ServiceSettingsVersion},
	{"storage", StorageSettingsVersion},
	{"networks", NetworkSettingsVersion},
}

// What swarm writes of a service on its own - its version, a rolling update
// marked complete - is no screen's: each saves as it was loaded.
func TestASettingsScreenIsNotMovedBySwarmWritingTheService(t *testing.T) {
	for _, screen := range screens {
		service := serviceForVersions()
		before := screen.version(service)
		service.Version.Index = 107
		service.UpdatedAt = time.Now()
		service.UpdateStatus = &swarm.UpdateStatus{State: swarm.UpdateStateCompleted, Message: "update completed"}
		assert.Equal(t, before, screen.version(service), screen.name)
	}
}

// Each screen's version moves with what it shows, and with nothing another
// screen shows.
func TestASettingsScreenIsMovedByWhatItShowsAlone(t *testing.T) {
	changes := map[string]func(*swarm.Service){
		"resources": func(s *swarm.Service) { s.Spec.TaskTemplate.Resources.Limits.MemoryBytes = 128 << 20 },
		"container": func(s *swarm.Service) { s.Spec.TaskTemplate.ContainerSpec.Image = "shop:2" },
		"availability": func(s *swarm.Service) {
			replicas := uint64(3)
			s.Spec.Mode.Replicated.Replicas = &replicas
		},
		"storage": func(s *swarm.Service) {
			s.Spec.TaskTemplate.ContainerSpec.Mounts = append(s.Spec.TaskTemplate.ContainerSpec.Mounts,
				mount.Mount{Type: mount.TypeVolume, Source: "logs", Target: "/logs"})
		},
		"networks": func(s *swarm.Service) {
			s.Spec.TaskTemplate.ContainerSpec.Hosts = []string{"10.0.0.2 db"}
		},
	}
	for changed, change := range changes {
		service := serviceForVersions()
		before := make(map[string]int, len(screens))
		for _, screen := range screens {
			before[screen.name] = screen.version(service)
		}
		change(service)
		for _, screen := range screens {
			if screen.name == changed {
				assert.NotEqual(t, before[screen.name], screen.version(service), "%s moves %s", changed, screen.name)
			} else {
				assert.Equal(t, before[screen.name], screen.version(service), "%s leaves %s", changed, screen.name)
			}
		}
	}
}

// The labels HivePaaS keeps - a domain's routing - are not the Container
// Settings screen's, nor is the Docker API socket the Persistent Storage
// screen's: neither moves them.
func TestWhatAScreenDoesNotChooseDoesNotMoveIt(t *testing.T) {
	service := serviceForVersions()
	container, storage := ContainerSettingsVersion(service), StorageSettingsVersion(service)

	service.Spec.Labels["traefik.http.routers.shop.rule"] = "Host(`shop.example.com`)"
	socket := mount.Mount{Type: mount.TypeVolume, Source: dockerapiservice.SocketVolumePrefix + "shop",
		Target: "/var/run/docker.sock"}
	service.Spec.TaskTemplate.ContainerSpec.Mounts = append(service.Spec.TaskTemplate.ContainerSpec.Mounts, socket)

	assert.Equal(t, container, ContainerSettingsVersion(service))
	assert.Equal(t, storage, StorageSettingsVersion(service))
	service.Spec.Labels["team"] = "store"
	assert.NotEqual(t, container, ContainerSettingsVersion(service), "a label of the operator's is the screen's")
}
