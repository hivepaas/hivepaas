package appprovisionserviceimpl

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/moby/moby/api/types/mount"
	"github.com/moby/moby/api/types/network"
	"github.com/moby/moby/api/types/swarm"
	"github.com/moby/moby/client"
	"github.com/stretchr/testify/assert"

	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/config"
	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/infra/database"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/bunex"
	"github.com/hivepaas/hivepaas/hivepaas_app/repository"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/appprovisionservice"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/appservice"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/clusterservice"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/networkservice"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/placementservice"
	"github.com/hivepaas/hivepaas/services/docker"
)

var errTestPersist = errors.New("database unavailable")

type fakeProjectRepo struct {
	repository.ProjectRepo
	project *entity.Project
}

func (f *fakeProjectRepo) GetByID(
	_ context.Context, _ database.IDB, _ string, _ ...bunex.SelectQueryOption,
) (*entity.Project, error) {
	return f.project, nil
}

type fakeAppRepo struct {
	repository.AppRepo
}

func (f *fakeAppRepo) GetByGlobalKey(
	_ context.Context, _ database.IDB, _, _ string, _ ...bunex.SelectQueryOption,
) (*entity.App, error) {
	return nil, hperrors.NewNotFound("App")
}

type fakeNetworkService struct {
	networkservice.Service
}

func (f *fakeNetworkService) GetOrCreateProjectNetwork(
	_ context.Context, _ database.IDB, _ *entity.Project, _ string,
) (*entity.Setting, *network.Inspect, error) {
	return nil, nil, nil
}

func (f *fakeNetworkService) GetProjectNetworkName(_ *entity.Project, env string) string {
	return "net-" + env
}

// fakePlacementService records the mounts it saw, to show placement runs after
// Configure has written them.
type fakePlacementService struct {
	placementservice.Service
	sawMounts int
}

func (f *fakePlacementService) ApplyPlacementSettings(
	_ context.Context, _ database.IDB, req *placementservice.ApplyPlacementSettingsReq,
) (*placementservice.ApplyPlacementSettingsResp, error) {
	f.sawMounts = len(req.Service.Spec.TaskTemplate.ContainerSpec.Mounts)
	return &placementservice.ApplyPlacementSettingsResp{Service: req.Service}, nil
}

type fakeDockerManager struct {
	docker.Manager
	created *swarm.ServiceSpec
}

func (f *fakeDockerManager) ServiceCreate(
	_ context.Context, spec *swarm.ServiceSpec, _ ...docker.ServiceCreateOption,
) (*client.ServiceCreateResult, error) {
	f.created = spec
	return &client.ServiceCreateResult{ID: "svc-1"}, nil
}

type fakeClusterService struct {
	clusterservice.Service
	removed []string
}

func (f *fakeClusterService) ServiceRemove(_ context.Context, serviceID string, _ int, _ time.Duration) error {
	f.removed = append(f.removed, serviceID)
	return nil
}

type fakeAppService struct {
	appservice.Service
	persisted *appservice.PersistingAppData
	err       error
}

func (f *fakeAppService) PersistAppData(
	_ context.Context, _ database.IDB, data *appservice.PersistingAppData,
) error {
	f.persisted = data
	return f.err
}

type provisionFakes struct {
	docker    *fakeDockerManager
	cluster   *fakeClusterService
	apps      *fakeAppService
	placement *fakePlacementService
}

func newProvisionTest(t *testing.T) (*service, *provisionFakes) {
	t.Helper()
	previous := config.Current()
	config.SetCurrent(&config.Config{Env: config.EnvProd})
	t.Cleanup(func() { config.SetCurrent(previous) })

	env := &entity.ProjectEnv{ID: "p1:prod", Key: "prod", Name: "prod", Status: base.ProjectStatusActive}
	project := &entity.Project{ID: "p1", Key: "shop", Name: "Shop", Status: base.ProjectStatusActive,
		ProjectEnvs: []*entity.ProjectEnv{env}}

	fakes := &provisionFakes{
		docker:    &fakeDockerManager{},
		cluster:   &fakeClusterService{},
		apps:      &fakeAppService{},
		placement: &fakePlacementService{},
	}
	svc := &service{
		dockerManager:    fakes.docker,
		appRepo:          &fakeAppRepo{},
		projectRepo:      &fakeProjectRepo{project: project},
		appService:       fakes.apps,
		clusterService:   fakes.cluster,
		networkService:   &fakeNetworkService{},
		placementService: fakes.placement,
	}
	return svc, fakes
}

func settingTypes(settings []*entity.Setting) []base.SettingType {
	types := make([]base.SettingType, 0, len(settings))
	for _, setting := range settings {
		types = append(types, setting.Type)
	}
	return types
}

func TestProvisionAppCreatesAnEmptyApp(t *testing.T) {
	svc, fakes := newProvisionTest(t)

	resp, err := svc.ProvisionApp(context.Background(), nil, &appprovisionservice.ProvisionAppReq{
		ProjectID: "p1", ProjectEnvID: "p1:prod", Name: "Web API", Status: base.AppStatusActive,
		Tags: []string{"api"},
	})

	assert.NoError(t, err)
	app := resp.App
	assert.Equal(t, "svc-1", app.ServiceID)
	assert.Equal(t, "shop", app.Project.Key)
	assert.Equal(t, "prod", app.ProjectEnv.Key)
	assert.Equal(t, []base.SettingType{base.SettingTypeAppRouting, base.SettingTypeAppFeatures},
		settingTypes(app.Settings))

	assert.Equal(t, dockerImageInit, fakes.docker.created.TaskTemplate.ContainerSpec.Image)
	assert.Equal(t, app.GlobalKey, fakes.docker.created.Name)
	assert.Equal(t, "net-prod", fakes.docker.created.TaskTemplate.Networks[0].Target)

	assert.Equal(t, []*entity.App{app}, fakes.apps.persisted.UpsertingApps)
	assert.Equal(t, app.Settings, fakes.apps.persisted.UpsertingSettings)
	assert.Equal(t, "api", fakes.apps.persisted.UpsertingTags[0].Tag)
	assert.Empty(t, fakes.cluster.removed)
}

func TestProvisionAppAppliesTheConfiguration(t *testing.T) {
	svc, fakes := newProvisionTest(t)
	routing := &entity.Setting{ID: "set-routing", Type: base.SettingTypeAppRouting}
	kind := &entity.Setting{ID: "set-kind", Type: base.SettingTypeAppKind}

	resp, err := svc.ProvisionApp(context.Background(), nil, &appprovisionservice.ProvisionAppReq{
		ProjectID: "p1", ProjectEnvID: "p1:prod", Name: "db", Status: base.AppStatusActive,
		Configure: func(_ context.Context, _ database.IDB, app *entity.App, spec *swarm.ServiceSpec) (
			[]*entity.Setting, error) {
			assert.NotEmpty(t, app.ID, "the app has its id before it is configured")
			spec.TaskTemplate.ContainerSpec.Mounts = append(spec.TaskTemplate.ContainerSpec.Mounts,
				mountAt("/data"))
			return []*entity.Setting{routing, kind}, nil
		},
	})

	assert.NoError(t, err)
	assert.Equal(t,
		[]base.SettingType{base.SettingTypeAppFeatures, base.SettingTypeAppRouting, base.SettingTypeAppKind},
		settingTypes(resp.App.Settings), "a configured setting replaces the default of its type")
	assert.Same(t, routing, resp.App.Settings[1])
	assert.Equal(t, 1, fakes.placement.sawMounts, "placement runs after Configure wrote the mounts")
	assert.Len(t, fakes.docker.created.TaskTemplate.ContainerSpec.Mounts, 1)
}

func TestProvisionAppCreatesNoServiceWhenConfigureFails(t *testing.T) {
	svc, fakes := newProvisionTest(t)

	_, err := svc.ProvisionApp(context.Background(), nil, &appprovisionservice.ProvisionAppReq{
		ProjectID: "p1", ProjectEnvID: "p1:prod", Name: "db", Status: base.AppStatusActive,
		Configure: func(context.Context, database.IDB, *entity.App, *swarm.ServiceSpec) ([]*entity.Setting, error) {
			return nil, hperrors.Wrap(hperrors.ErrSpecBlockInvalid)
		},
	})

	assert.ErrorIs(t, err, hperrors.ErrSpecBlockInvalid)
	assert.Nil(t, fakes.docker.created)
}

func TestProvisionAppRemovesTheServiceWhenPersistingFails(t *testing.T) {
	svc, fakes := newProvisionTest(t)
	fakes.apps.err = errTestPersist

	_, err := svc.ProvisionApp(context.Background(), nil, &appprovisionservice.ProvisionAppReq{
		ProjectID: "p1", ProjectEnvID: "p1:prod", Name: "db", Status: base.AppStatusActive,
	})

	assert.ErrorIs(t, err, errTestPersist)
	assert.Equal(t, []string{"svc-1"}, fakes.cluster.removed)
}

func TestProvisionAppRefusesAnInactiveEnv(t *testing.T) {
	svc, fakes := newProvisionTest(t)
	svc.projectRepo.(*fakeProjectRepo).project.ProjectEnvs[0].Status = base.ProjectStatusDisabled

	_, err := svc.ProvisionApp(context.Background(), nil, &appprovisionservice.ProvisionAppReq{
		ProjectID: "p1", ProjectEnvID: "p1:prod", Name: "db", Status: base.AppStatusActive,
	})

	assert.ErrorIs(t, err, hperrors.ErrProjectEnvInactive)
	assert.Nil(t, fakes.docker.created)
}

func mountAt(target string) mount.Mount {
	return mount.Mount{Type: mount.TypeVolume, Source: "vol", Target: target}
}
