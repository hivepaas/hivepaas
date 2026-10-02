package traefikservice

import (
	"testing"

	"github.com/moby/moby/api/types/swarm"
	"github.com/stretchr/testify/assert"
)

func traefikSpec(args ...string) *swarm.ServiceSpec {
	return &swarm.ServiceSpec{TaskTemplate: swarm.TaskSpec{
		ContainerSpec: &swarm.ContainerSpec{Args: args, Labels: map[string]string{"keep": "me"}},
		LogDriver: &swarm.Driver{Name: "json-file",
			Options: map[string]string{"max-size": "50m", "labels": "other.label"}},
	}}
}

// The proxy's lines get its identity, its other labels and options kept; a
// second time changes nothing, so the spec stays the same.
func TestWithAccessLogIdentity(t *testing.T) {
	spec := traefikSpec("--accesslog=true", "--accesslog.format=json")
	assert.Equal(t, AccessLogUnlabelled, AccessLogReadiness(spec))

	assert.True(t, WithAccessLogIdentity(spec))
	cs := spec.TaskTemplate.ContainerSpec
	assert.Equal(t, "traefik", cs.Labels["hivepaas.component"])
	assert.Equal(t, "me", cs.Labels["keep"])
	assert.Equal(t, "hivepaas.component,other.label", spec.TaskTemplate.LogDriver.Options["labels"])
	assert.Equal(t, "50m", spec.TaskTemplate.LogDriver.Options["max-size"])
	assert.Equal(t, AccessLogNotReadyReason(""), AccessLogReadiness(spec))

	assert.False(t, WithAccessLogIdentity(spec), "already there")

	noDriver := traefikSpec()
	noDriver.TaskTemplate.LogDriver = nil
	assert.True(t, WithAccessLogIdentity(noDriver))
	assert.Equal(t, "json-file", noDriver.TaskTemplate.LogDriver.Name)
	assert.Equal(t, "hivepaas.component", noDriver.TaskTemplate.LogDriver.Options["labels"])

	other := traefikSpec()
	other.TaskTemplate.LogDriver = &swarm.Driver{Name: "syslog"}
	WithAccessLogIdentity(other)
	assert.Equal(t, "syslog", other.TaskTemplate.LogDriver.Name, "another driver is the operator's")
}

// Readiness says what is missing, first things first.
func TestAccessLogReadiness(t *testing.T) {
	ready := traefikSpec("traefik", "--accesslog=true", "--accesslog.format=json")
	WithAccessLogIdentity(ready)
	assert.Equal(t, AccessLogNotReadyReason(""), AccessLogReadiness(ready))

	off := traefikSpec("--log.level=INFO")
	WithAccessLogIdentity(off)
	assert.Equal(t, AccessLogOff, AccessLogReadiness(off))

	clf := traefikSpec("--accesslog=true")
	WithAccessLogIdentity(clf)
	assert.Equal(t, AccessLogNotJSON, AccessLogReadiness(clf))

	assert.Equal(t, AccessLogOff, AccessLogReadiness(nil))
}
