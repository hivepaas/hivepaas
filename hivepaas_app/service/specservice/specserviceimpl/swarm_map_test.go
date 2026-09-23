package specserviceimpl

import (
	"testing"

	"github.com/moby/moby/api/types/mount"
	"github.com/moby/moby/api/types/swarm"
	"github.com/stretchr/testify/assert"
)

func testService() *swarm.Service {
	return &swarm.Service{
		Spec: swarm.ServiceSpec{
			Annotations: swarm.Annotations{
				Labels: map[string]string{
					"hivepaas.app.info":                 `{"name":"a1"}`,
					"hivepaas.app.placementConstraints": "node.role==manager",
					"team":                              "platform",
				},
			},
			TaskTemplate: swarm.TaskSpec{
				ContainerSpec: &swarm.ContainerSpec{
					Image: "ghcr.io/acme/api:1.4.2@sha256:" +
						"6ecdf4e6779ce1a655b53dcbab6d8920d971d7e86aab9a2186ae43fbc9df4dfb",
					Hostname: "api",
					Mounts: []mount.Mount{
						{Type: mount.TypeVolume, Source: "vol_1", Target: "/var/lib/postgresql/data"},
						{Type: mount.TypeBind, Source: "/srv/conf", Target: "/etc/app/config"},
						{
							Type: mount.TypeVolume, Source: "hp-shared", Target: "/shared",
							VolumeOptions: &mount.VolumeOptions{Subpath: "project_a/dev/backend/cache"},
						},
						{
							Type: mount.TypeTmpfs, Target: "/dev/shm",
							TmpfsOptions: &mount.TmpfsOptions{SizeBytes: 64 << 20},
						},
					},
				},
				Networks: []swarm.NetworkAttachmentConfig{
					{Target: "8vo4p3pwm1aksdu2ilryn8mpf", Aliases: []string{"api"}},
				},
				Placement: &swarm.Placement{
					Constraints: []string{"node.role == manager", "node.labels.zone == eu"},
					Platforms:   []swarm.Platform{{Architecture: "amd64", OS: "linux"}},
				},
			},
		},
	}
}

func TestMapSwarmServiceStripsTheImageDigest(t *testing.T) {
	out := mapSwarmService(testService(), nil)
	assert.Equal(t, "ghcr.io/acme/api:1.4.2", out.Container.Image,
		"a digest pins an image that may not exist in the target registry")
}

func TestMapSwarmServiceKeepsAnImageWithNoDigest(t *testing.T) {
	svc := testService()
	svc.Spec.TaskTemplate.ContainerSpec.Image = "crccheck/hello-world:latest"
	out := mapSwarmService(svc, nil)
	assert.Equal(t, "crccheck/hello-world:latest", out.Container.Image)
}

func TestMapSwarmServiceResolvesNetworkNames(t *testing.T) {
	out := mapSwarmService(testService(), map[string]string{
		"8vo4p3pwm1aksdu2ilryn8mpf": "p1_dev_net",
	})
	assert.Len(t, out.Networks.Attachments, 1)
	assert.Equal(t, "p1_dev_net", out.Networks.Attachments[0].Name)
	assert.Equal(t, []string{"api"}, out.Networks.Attachments[0].Aliases)
}

// An id with no name is a network the exporter could not resolve. It is kept as
// the id rather than dropped, so import can report it rather than lose the
// attachment silently.
func TestMapSwarmServiceKeepsUnresolvedNetworkIDs(t *testing.T) {
	out := mapSwarmService(testService(), nil)
	assert.Equal(t, "8vo4p3pwm1aksdu2ilryn8mpf", out.Networks.Attachments[0].Name)
}

func TestMapSwarmServiceDropsDerivedPlacement(t *testing.T) {
	out := mapSwarmService(testService(), nil)
	assert.Equal(t, []string{"node.labels.zone == eu"}, out.Service.Placement.Constraints,
		"node.role == manager was added by HivePaaS and is regenerated")
}

func TestMapSwarmServiceKeepsOnlyUserLabels(t *testing.T) {
	out := mapSwarmService(testService(), nil)
	assert.Equal(t, map[string]string{"team": "platform"}, out.Container.ServiceLabels)
}

func TestMapSwarmServiceHandlesAServiceWithNoContainerSpec(t *testing.T) {
	svc := testService()
	svc.Spec.TaskTemplate.ContainerSpec = nil
	out := mapSwarmService(svc, nil)
	assert.Nil(t, out.Container)
	assert.Nil(t, out.Storage)
}

func TestMapSwarmServiceIsNilForNoService(t *testing.T) {
	out := mapSwarmService(nil, nil)
	assert.Nil(t, out)
}
