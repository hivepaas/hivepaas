package appsettingsuc

import (
	"context"
	"testing"
	"time"

	"github.com/moby/moby/api/types/mount"
	"github.com/moby/moby/api/types/swarm"
	"github.com/stretchr/testify/assert"

	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/basedto"
	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/infra/database"
	"github.com/hivepaas/hivepaas/hivepaas_app/permission"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/bunex"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/appservice"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/placementservice"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/volumeservice"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/appsettingsuc/appsettingsdto"
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

// The service settings form owns the constraint list, and saving it writes that
// list wholesale - which drops the ones HivePaaS manages: volume pins and the
// placement settings. They have to be put back in the same spec, or the app is
// free to schedule anywhere until something else rebuilds them.
func TestApplyAppServiceSettingsRestoresManagedConstraints(t *testing.T) {
	dockerManager := &serviceUpdatingDockerManager{}
	placement := &recordingPlacementService{}
	uc := &UC{dockerManager: dockerManager, placementService: placement}

	data := &updateAppServiceSettingsData{
		App: &entity.App{ID: "app-1", ServiceID: "svc-1"},
		Service: &swarm.Service{
			ID: "svc-1",
			Spec: swarm.ServiceSpec{
				TaskTemplate: swarm.TaskSpec{
					ContainerSpec: &swarm.ContainerSpec{},
					// What HivePaaS had put there before this save.
					Placement: &swarm.Placement{Constraints: []string{"node.id==node-2"}},
				},
			},
		},
	}
	// A request with no placement of its own: the form sent none, so the
	// preparation clears the list entirely.
	req := &appsettingsdto.UpdateAppServiceSettingsReq{}

	if err := uc.applyAppServiceSettings(context.Background(), database.Tx{}, req, data); err != nil {
		t.Fatalf("applyAppServiceSettings failed: %v", err)
	}

	if dockerManager.updatedSpec == nil {
		t.Fatal("the settings must actually have been written")
	}
	if dockerManager.updatedSpec.TaskTemplate.Placement == nil {
		t.Fatal("the spec reached docker with no placement at all")
	}
	assert.Equal(t, []string{"node.id==node-2"},
		dockerManager.updatedSpec.TaskTemplate.Placement.Constraints,
		"the managed constraint has to be back in the spec that rolls the tasks")

	if len(placement.reqs) != 1 {
		t.Fatalf("expected one apply, got %d", len(placement.reqs))
	}
	assert.True(t, placement.reqs[0].SkipSavingToDocker,
		"placementservice must mutate this spec, not save a second one")
	assert.Equal(t, data.App, placement.reqs[0].App)
}

// ownerLookup answers the two questions resolving an owner asks: which app the
// mount names, and whether the caller may have its data.
type ownerLookupAppService struct {
	appservice.Service
	apps map[string]*entity.App
}

func (f *ownerLookupAppService) LoadApp(
	_ context.Context, _ database.IDB, _, appID string, _, _ bool, _ ...bunex.SelectQueryOption,
) (*entity.App, error) {
	app, found := f.apps[appID]
	if !found {
		return nil, hperrors.NewNotFound("App")
	}
	return app, nil
}

type answeringPermissionManager struct {
	permission.Manager
	allow   bool
	checked []permission.AccessCheck
}

func (f *answeringPermissionManager) CheckAccess(
	_ context.Context, _ database.IDB, _ *basedto.Auth, check permission.AccessCheck,
) (bool, error) {
	f.checked = append(f.checked, check)
	return f.allow, nil
}

func ownerResolutionTest(allow bool, owner *entity.App) (*UC, *answeringPermissionManager) {
	perms := &answeringPermissionManager{allow: allow}
	return &UC{
		appService:        &ownerLookupAppService{apps: map[string]*entity.App{owner.ID: owner}},
		permissionManager: perms,
	}, perms
}

func mountingApp() *entity.App {
	return &entity.App{ID: "app-1", Key: "files", ProjectID: "proj-1", ProjectEnvID: "env-1"}
}

func ownerApp() *entity.App {
	return &entity.App{
		ID: "app-2", Key: "postgres", Name: "Postgres", ProjectID: "proj-1", ProjectEnvID: "env-1",
		ProjectEnv: &entity.ProjectEnv{Key: "prod"},
	}
}

// The grant is everything in the owner's directory, so it is checked against the
// owner rather than against the project.
func TestResolveMountOwnerAppChecksTheOwner(t *testing.T) {
	owner := ownerApp()
	uc, perms := ownerResolutionTest(true, owner)
	out := &volumeservice.AppMountReq{}

	err := uc.resolveMountOwnerApp(context.Background(), nil, nil, mountingApp(),
		&appsettingsdto.MountSourceApp{AppID: owner.ID, Write: true}, out)

	assert.NoError(t, err)
	assert.Equal(t, owner, out.OwnerApp)
	assert.False(t, out.ReadOnly, "write was asked for")
	assert.Len(t, perms.checked, 1)
	appCheck, ok := perms.checked[0].(*permission.AppAccessCheck)
	assert.True(t, ok)
	assert.Equal(t, owner.ID, appCheck.AppID)
	assert.Equal(t, base.ActionTypeWrite, appCheck.Action)
}

// Seeing another app's files is one decision and changing them is another: a
// request that says nothing has only asked for the first.
func TestResolveMountOwnerAppIsReadOnlyUnlessWriteIsAsked(t *testing.T) {
	owner := ownerApp()
	uc, _ := ownerResolutionTest(true, owner)
	out := &volumeservice.AppMountReq{ReadOnly: false}

	err := uc.resolveMountOwnerApp(context.Background(), nil, nil, mountingApp(),
		&appsettingsdto.MountSourceApp{AppID: owner.ID}, out)

	assert.NoError(t, err)
	assert.True(t, out.ReadOnly, "the mount's own readOnly does not override the answer given here")
}

func TestResolveMountOwnerAppRefusesWithoutPermission(t *testing.T) {
	owner := ownerApp()
	uc, _ := ownerResolutionTest(false, owner)
	out := &volumeservice.AppMountReq{}

	err := uc.resolveMountOwnerApp(context.Background(), nil, nil, mountingApp(),
		&appsettingsdto.MountSourceApp{AppID: owner.ID}, out)

	assert.ErrorIs(t, err, hperrors.ErrUnauthorized)
	assert.Nil(t, out.OwnerApp)
}

// An app may only be given the storage of an app beside it: another environment
// is a different conversation, and the data there is not this one's.
func TestResolveMountOwnerAppRefusesAnotherEnvironment(t *testing.T) {
	owner := ownerApp()
	owner.ProjectEnvID = "env-2"
	uc, perms := ownerResolutionTest(true, owner)
	out := &volumeservice.AppMountReq{}

	err := uc.resolveMountOwnerApp(context.Background(), nil, nil, mountingApp(),
		&appsettingsdto.MountSourceApp{AppID: owner.ID}, out)

	assert.Error(t, err)
	assert.Empty(t, perms.checked, "refused before anything is asked of the permission manager")
}

// Its own directory is the ordinary case, and asks nobody anything.
func TestResolveMountOwnerAppIgnoresItself(t *testing.T) {
	app := mountingApp()
	uc, perms := ownerResolutionTest(true, ownerApp())
	out := &volumeservice.AppMountReq{}

	for name, src := range map[string]*appsettingsdto.MountSourceApp{
		"no source app": nil,
		"an empty id":   {AppID: ""},
		"its own id":    {AppID: app.ID},
	} {
		t.Run(name, func(t *testing.T) {
			assert.NoError(t, uc.resolveMountOwnerApp(context.Background(), nil, nil, app, src, out))
			assert.Nil(t, out.OwnerApp)
			assert.Empty(t, perms.checked)
		})
	}
}
