package appsettingsdto

import (
	"testing"

	"github.com/moby/moby/api/types/swarm"
	"github.com/stretchr/testify/assert"
)

func labeledService() *swarm.Service {
	return &swarm.Service{Spec: swarm.ServiceSpec{
		Annotations: swarm.Annotations{Labels: map[string]string{
			"com.docker.stack.namespace": "shop_prod",
			"hivepaas.app.id":            "app-1",
			"team":                       "web",
		}},
		TaskTemplate: swarm.TaskSpec{ContainerSpec: &swarm.ContainerSpec{Labels: map[string]string{
			"com.docker.stack.image": "nginx:1.27",
			"traefik.enable":         "false",
			"tier":                   "front",
		}}},
	}}
}

// The labels HivePaaS and Docker manage are left out unless they are asked for:
// changing them on this screen does nothing, and they are not the person's.
func TestContainerSettingsLeaveOutSystemLabels(t *testing.T) {
	resp, err := TransformContainerSettings(labeledService(), false)

	assert.NoError(t, err)
	assert.Equal(t, map[string]string{"team": "web"}, resp.ServiceLabels)
	assert.Equal(t, map[string]string{"tier": "front"}, resp.ContainerLabels)
}

func TestContainerSettingsRevealSystemLabelsWhenAsked(t *testing.T) {
	resp, err := TransformContainerSettings(labeledService(), true)

	assert.NoError(t, err)
	assert.Len(t, resp.ServiceLabels, 3)
	assert.Len(t, resp.ContainerLabels, 3)
}
