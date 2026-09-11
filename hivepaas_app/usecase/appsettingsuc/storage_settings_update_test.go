package appsettingsuc

import (
	"context"
	"testing"
	"time"

	"github.com/moby/moby/api/types/mount"
	"github.com/moby/moby/api/types/swarm"
	"github.com/stretchr/testify/assert"

	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/infra/database"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/placementservice"
	"github.com/hivepaas/hivepaas/services/docker"
)

// serviceUpdatingDockerManager runs the update callback once against the spec it
// was handed and records what the callback left behind, which is the spec that
// would have gone to docker.
type serviceUpdatingDockerManager struct {
	docker.Manager

	updatedSpec *swarm.ServiceSpec
}

func (m *serviceUpdatingDockerManager) ServiceUpdateFunc(
	_ context.Context,
	_ string,
	initialService *swarm.Service,
	fn func(int, *swarm.Service) (bool, error),
	_ int,
	_ time.Duration,
	_ ...docker.ServiceUpdateOption,
) error {
	ok, err := fn(0, initialService)
	if err != nil || !ok {
		return err
	}
	m.updatedSpec = &initialService.Spec
	return nil
}

// recordingPlacementService stands in for the real one and writes a constraint
// the test can look for, so the assertion is about the spec that reached docker
// rather than about the call having been made.
type recordingPlacementService struct {
	reqs []*placementservice.ApplyPlacementSettingsReq
}

func (s *recordingPlacementService) ApplyPlacementSettings(
	_ context.Context,
	_ database.IDB,
	req *placementservice.ApplyPlacementSettingsReq,
) (*placementservice.ApplyPlacementSettingsResp, error) {
	s.reqs = append(s.reqs, req)
	req.Service.Spec.TaskTemplate.Placement = &swarm.Placement{Constraints: []string{"node.id==node-2"}}
	return &placementservice.ApplyPlacementSettingsResp{Service: req.Service, HasChanges: true}, nil
}

// Writing the mounts rolls the service's tasks. If the pin constraint waits for
// the next deploy, a task rescheduled by this very update can land on a node
// holding none of the data - the failure the pin exists to prevent. So the
// constraint has to be in the same spec as the mounts, in one update.
func TestApplyAppStorageSettingsPutsThePinConstraintInTheSameUpdateAsTheMounts(t *testing.T) {
	dockerManager := &serviceUpdatingDockerManager{}
	placement := &recordingPlacementService{}
	uc := &UC{dockerManager: dockerManager, placementService: placement}

	data := &updateAppStorageSettingsData{
		App: &entity.App{ID: "app-1", ServiceID: "svc-1"},
		Service: &swarm.Service{
			ID: "svc-1",
			Spec: swarm.ServiceSpec{
				TaskTemplate: swarm.TaskSpec{ContainerSpec: &swarm.ContainerSpec{}},
			},
		},
		FinalMounts: []mount.Mount{{Type: mount.TypeBind, Source: "/srv/data/web", Target: "/data"}},
	}

	if err := uc.applyAppStorageSettings(context.Background(), nil, data); err != nil {
		t.Fatalf("applyAppStorageSettings failed: %v", err)
	}

	if dockerManager.updatedSpec == nil {
		t.Fatal("the mounts must actually have been written")
	}
	assert.Equal(t, data.FinalMounts, dockerManager.updatedSpec.TaskTemplate.ContainerSpec.Mounts)
	if dockerManager.updatedSpec.TaskTemplate.Placement == nil {
		t.Fatal("expected a placement to have been set on the updated spec")
	}
	assert.Equal(t, []string{"node.id==node-2"},
		dockerManager.updatedSpec.TaskTemplate.Placement.Constraints,
		"the constraint has to ride along in the spec that rolls the tasks")

	if len(placement.reqs) != 1 {
		t.Fatalf("expected one update, not a second one that rolls the tasks again: got %d", len(placement.reqs))
	}
	assert.True(t, placement.reqs[0].SkipSavingToDocker,
		"placementservice must mutate this spec, not save one of its own")
	assert.Equal(t, data.App, placement.reqs[0].App)
}
