package appautoscaleserviceimpl

import (
	"context"
	"testing"

	"github.com/moby/moby/api/types/swarm"
	"github.com/stretchr/testify/assert"
)

// A service's pending tasks are those wanted and not running; a service Docker
// does not answer has none.
func TestPendingIsTheTasksWantedAndNotRunning(t *testing.T) {
	s := &service{dockerManager: &fakeSwarm{services: map[string]*swarm.Service{
		"svc1": {ID: "svc1", ServiceStatus: &swarm.ServiceStatus{DesiredTasks: 3, RunningTasks: 1}},
		"svc2": {ID: "svc2", ServiceStatus: &swarm.ServiceStatus{DesiredTasks: 2, RunningTasks: 2}},
	}}}
	for id, want := range map[string]int{"svc1": 2, "svc2": 0, "gone": 0} {
		got, err := s.pending(context.Background(), id)
		assert.NoError(t, err, id)
		assert.Equal(t, want, got, id)
	}
}
