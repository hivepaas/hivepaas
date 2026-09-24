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
	"github.com/hivepaas/hivepaas/hivepaas_app/service/appdeploymentservice"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/appprovisionservice"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/approutingservice"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/appservice"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/clustersecretservice"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/clusterservice"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/domainservice"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/envvarservice"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/networkservice"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/placementservice"
	"github.com/hivepaas/hivepaas/services/docker"
)

var (
	errTestPersist   = errors.New("database unavailable")
	errTestRouting   = errors.New("traefik unavailable")
	errTestProvision = errors.New("provisioning failed")
)

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

// fakeEnvVarService records that the environment was rebuilt.
type fakeEnvVarService struct {
	envvarservice.Service
	applied bool
}

func (f *fakeEnvVarService) BuildEnvVarsForAllAppsInScope(
	context.Context, database.IDB, *entity.ObjectScope, bool, []string, bool, bool,
) ([]*envvarservice.AppEnvVarData, error) {
	return nil, nil
}

func (f *fakeEnvVarService) ApplyEnvVarsForApps(
	context.Context, database.IDB, []*envvarservice.AppEnvVarData, bool, bool,
) map[int]error {
	f.applied = true
	return nil
}

type fakeRoutingService struct {
	approutingservice.Service
	req *approutingservice.ApplyAppRoutingReq
	err error
	// failTimes is how many of the first calls fail with err before it succeeds;
	// err alone fails every call.
	failTimes int
	calls     int
}

func (f *fakeRoutingService) ApplyRoutingSettings(
	_ context.Context, _ database.IDB, req *approutingservice.ApplyAppRoutingReq,
) (*approutingservice.ApplyAppRoutingResp, error) {
	f.req = req
	f.calls++
	if f.err != nil && (f.failTimes == 0 || f.calls <= f.failTimes) {
		return nil, f.err
	}
	return &approutingservice.ApplyAppRoutingResp{}, nil
}

// fakeClusterSecretService fills in what docker would have returned: an id for
// every entry that asked to be mounted as a file, and nothing for the rest.
type fakeClusterSecretService struct {
	clustersecretservice.Service
	secrets        []*entity.Secret
	configs        []*entity.ConfigFile
	removedSecrets []string
	removedConfigs []string
}

func (f *fakeClusterSecretService) CreateSecretsForApp(
	_ context.Context, _ database.IDB, _ *entity.App, secrets []*entity.Secret,
) ([]*entity.SwarmSecretRef, error) {
	f.secrets = secrets
	refs := make([]*entity.SwarmSecretRef, 0, len(secrets))
	for _, secret := range secrets {
		if secret.SwarmRef == nil || secret.SwarmRef.File == nil {
			refs = append(refs, nil)
			continue
		}
		secret.SwarmRef.SecretID = "docker-secret-" + secret.Key
		refs = append(refs, secret.SwarmRef)
	}
	return refs, nil
}

func (f *fakeClusterSecretService) CreateConfigsForApp(
	_ context.Context, _ database.IDB, _ *entity.App, configs []*entity.ConfigFile,
) ([]*entity.SwarmConfigRef, error) {
	f.configs = configs
	refs := make([]*entity.SwarmConfigRef, 0, len(configs))
	for _, config := range configs {
		if config.SwarmRef == nil || config.SwarmRef.File == nil {
			refs = append(refs, nil)
			continue
		}
		config.SwarmRef.ConfigID = "docker-config-" + config.Name
		refs = append(refs, config.SwarmRef)
	}
	return refs, nil
}

func (f *fakeClusterSecretService) SecretsRemove(
	_ context.Context, secretIDs []string, _ int, _ time.Duration,
) error {
	f.removedSecrets = append(f.removedSecrets, secretIDs...)
	return nil
}

func (f *fakeClusterSecretService) ConfigsRemove(
	_ context.Context, configIDs []string, _ int, _ time.Duration,
) error {
	f.removedConfigs = append(f.removedConfigs, configIDs...)
	return nil
}

type fakeDeploymentService struct {
	appdeploymentservice.Service
}

func (f *fakeDeploymentService) CreateDeploymentAndTask(
	app *entity.App, settings *entity.AppDeploymentSettings, _ appdeploymentservice.DeploymentArgs,
) (*entity.Deployment, *entity.Task, error) {
	return &entity.Deployment{ID: "dep-1", AppID: app.ID, Settings: settings},
		&entity.Task{ID: "task-1"}, nil
}

// fakeDomainService hands back whatever certificate a test put in it, keyed by
// the domain it covers.
type fakeDomainService struct {
	domainservice.Service
	certs     map[string]*entity.Setting
	asked     []string
	obtaining []string
}

func (f *fakeDomainService) FindCertsForDomains(
	_ context.Context, _ database.IDB, _ *entity.ObjectScope, domains []string,
) (map[string]*entity.Setting, error) {
	f.asked = append(f.asked, domains...)
	found := map[string]*entity.Setting{}
	for _, domain := range domains {
		if cert := f.certs[domain]; cert != nil {
			found[domain] = cert
		}
	}
	return found, nil
}

func (f *fakeDomainService) EnsureCertsForDomains(
	ctx context.Context, db database.IDB, req *domainservice.EnsureCertsReq,
) (*domainservice.EnsureCertsResp, error) {
	matched, err := f.FindCertsForDomains(ctx, db, req.Scope, req.Domains)
	if err != nil {
		return nil, err
	}
	resp := &domainservice.EnsureCertsResp{Matched: matched, Skipped: map[string]string{}}
	for _, domain := range req.Domains {
		if matched[domain] == nil {
			f.obtaining = append(f.obtaining, domain)
		}
	}
	return resp, nil
}

type provisionFakes struct {
	docker       *fakeDockerManager
	cluster      *fakeClusterService
	apps         *fakeAppService
	placement    *fakePlacementService
	envVars      *fakeEnvVarService
	routing      *fakeRoutingService
	clusterFiles *fakeClusterSecretService
	domains      *fakeDomainService
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
		docker:       &fakeDockerManager{},
		cluster:      &fakeClusterService{},
		apps:         &fakeAppService{},
		placement:    &fakePlacementService{},
		envVars:      &fakeEnvVarService{},
		routing:      &fakeRoutingService{},
		clusterFiles: &fakeClusterSecretService{},
		domains:      &fakeDomainService{},
	}
	svc := &service{
		dockerManager:        fakes.docker,
		appRepo:              &fakeAppRepo{},
		projectRepo:          &fakeProjectRepo{project: project},
		appDeploymentService: &fakeDeploymentService{},
		appRoutingService:    fakes.routing,
		appService:           fakes.apps,
		clusterSecretService: fakes.clusterFiles,
		clusterService:       fakes.cluster,
		domainService:        fakes.domains,
		envVarService:        fakes.envVars,
		networkService:       &fakeNetworkService{},
		placementService:     fakes.placement,
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

// A key given is kept rather than derived from the name: an imported app keeps
// the key it had, whatever its name says now.
func TestProvisionAppKeepsAGivenKey(t *testing.T) {
	svc, _ := newProvisionTest(t)

	resp, err := svc.ProvisionApp(context.Background(), nil, &appprovisionservice.ProvisionAppReq{
		ProjectID: "p1", ProjectEnvID: "p1:prod", Key: "api", Name: "Web API", Status: base.AppStatusActive,
	})

	assert.NoError(t, err)
	assert.Equal(t, "api", resp.App.Key)
	assert.Equal(t, "Web API", resp.App.Name)
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
