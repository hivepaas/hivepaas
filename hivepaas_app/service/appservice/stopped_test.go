package appservice

import (
	"testing"

	"github.com/moby/moby/api/types/swarm"
	"github.com/stretchr/testify/assert"
)

func TestStopAndStartGlobalService(t *testing.T) {
	spec := &swarm.ServiceSpec{TaskTemplate: swarm.TaskSpec{
		Placement: &swarm.Placement{Constraints: []string{"node.labels.tier==gold"}},
	}}

	assert.True(t, StopGlobalService(spec))
	assert.Equal(t, []string{"node.labels.tier==gold", ConstraintAppStopped}, spec.TaskTemplate.Placement.Constraints)
	assert.False(t, StopGlobalService(spec), "stopping twice changes nothing")

	assert.True(t, StartGlobalService(spec))
	assert.Equal(t, []string{"node.labels.tier==gold"}, spec.TaskTemplate.Placement.Constraints)
	assert.False(t, StartGlobalService(spec), "starting a running service changes nothing")
}

func TestStopGlobalServiceWithoutPlacement(t *testing.T) {
	spec := &swarm.ServiceSpec{}

	assert.False(t, StartGlobalService(spec))
	assert.True(t, StopGlobalService(spec))
	assert.Equal(t, []string{ConstraintAppStopped}, spec.TaskTemplate.Placement.Constraints)
}

// A spec about to become another service carries no stopped state into it.
func TestForgetStoppedState(t *testing.T) {
	spec := &swarm.ServiceSpec{
		Annotations: swarm.Annotations{Labels: map[string]string{LabelAppPrevServiceMode: "{}", "keep": "1"}},
		TaskTemplate: swarm.TaskSpec{
			Placement: &swarm.Placement{Constraints: []string{ConstraintAppStopped}},
		},
	}

	ForgetStoppedState(spec)

	assert.Equal(t, map[string]string{"keep": "1"}, spec.Labels)
	assert.Empty(t, spec.TaskTemplate.Placement.Constraints)
}
