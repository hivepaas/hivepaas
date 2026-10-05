package appuc

import (
	"context"
	"testing"

	"github.com/moby/moby/api/types/swarm"
	"github.com/moby/moby/client"
	"github.com/stretchr/testify/assert"

	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/services/docker"
)

// fakeServices answers the services of the ids it is asked for, as the docker
// manager does, and keeps what it was asked.
type fakeServices struct {
	docker.Manager
	services map[string]swarm.Service
	asked    []string
	status   bool
}

func (f *fakeServices) ServiceListByIDs(_ context.Context, ids []string, options ...docker.ServiceListOption) (
	*client.ServiceListResult, error) {
	var opts client.ServiceListOptions
	for _, opt := range options {
		opt(&opts)
	}
	f.status = opts.Status
	out := &client.ServiceListResult{}
	for _, id := range ids {
		if id != "" {
			f.asked = append(f.asked, id)
		}
		if svc, ok := f.services[id]; ok {
			out.Items = append(out.Items, svc)
		}
	}
	return out, nil
}

// The services of the apps and their children come from one list of their ids,
// with task counts; an app never deployed, or whose service is gone, has none.
func TestAppServicesAreListedByTheirIDs(t *testing.T) {
	swarmFake := &fakeServices{services: map[string]swarm.Service{
		"svc-web":     {ID: "svc-web", ServiceStatus: &swarm.ServiceStatus{RunningTasks: 2, DesiredTasks: 2}},
		"svc-preview": {ID: "svc-preview"},
		"svc-other":   {ID: "svc-other"},
	}}
	web := &entity.App{ID: "web", ServiceID: "svc-web",
		ChildApps:        []*entity.App{{ID: "preview", ServiceID: "svc-preview"}},
		LogicalChildApps: []*entity.App{{ID: "db"}}}
	gone := &entity.App{ID: "gone", ServiceID: "svc-gone"}
	uc := &UC{dockerManager: swarmFake}

	got, err := uc.loadAppSwarmServices(context.Background(), []*entity.App{web, gone})
	if !assert.NoError(t, err) {
		t.FailNow()
	}
	assert.ElementsMatch(t, []string{"svc-web", "svc-preview", "svc-gone"}, swarmFake.asked)
	assert.True(t, swarmFake.status)
	assert.Len(t, got, 4) //nolint:mnd
	if assert.NotNil(t, got["web"]) {
		assert.Equal(t, uint64(2), got["web"].ServiceStatus.RunningTasks)
	}
	// Listed without task counts, a service is given zero ones.
	if assert.NotNil(t, got["preview"]) {
		assert.Equal(t, &swarm.ServiceStatus{}, got["preview"].ServiceStatus)
	}
	assert.Nil(t, got["db"])
	assert.Nil(t, got["gone"])
}
