package loggingserviceimpl

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	"github.com/moby/moby/api/types/network"
	"github.com/moby/moby/api/types/swarm"
	"github.com/moby/moby/client"

	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/infra/database"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/bunex"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/datakey"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/logging"
	"github.com/hivepaas/hivepaas/hivepaas_app/repository"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/systemappservice"
	"github.com/hivepaas/hivepaas/services/docker"
)

// fakeDocker holds the services the tests look at. Embedding the interface means
// only the methods these paths use need a body; anything else panics loudly
// instead of passing quietly.
type fakeDocker struct {
	docker.Manager
	inspected map[string]swarm.Service
	networks  []network.Summary
}

func (f *fakeDocker) ServiceInspect(
	_ context.Context, serviceID string, _ ...docker.ServiceInspectOption,
) (*client.ServiceInspectResult, error) {
	if svc, ok := f.inspected[serviceID]; ok {
		return &client.ServiceInspectResult{Service: svc}, nil
	}
	return nil, hperrors.Wrap(hperrors.ErrNotFound)
}

// ServiceList returns every service; the code under test picks its own by id.
func (f *fakeDocker) ServiceList(
	_ context.Context, _ ...docker.ServiceListOption,
) (*client.ServiceListResult, error) {
	items := make([]swarm.Service, 0, len(f.inspected))
	for _, svc := range f.inspected {
		items = append(items, svc)
	}
	return &client.ServiceListResult{Items: items}, nil
}

// ServiceUpdateFunc mirrors what the manager does: hand the callback the
// service, respect its "nothing to do" answer, and only then store the change.
func (f *fakeDocker) ServiceUpdateFunc(
	_ context.Context, serviceID string, service *swarm.Service,
	fn func(int, *swarm.Service) (bool, error), _ int, _ time.Duration,
	_ ...docker.ServiceUpdateOption,
) error {
	if service == nil {
		current, ok := f.inspected[serviceID]
		if !ok {
			return hperrors.Wrap(hperrors.ErrNotFound)
		}
		service = &current
	}
	apply, err := fn(0, service)
	if err != nil || !apply {
		return err
	}
	f.inspected[serviceID] = *service
	return nil
}

func (f *fakeDocker) NetworkList(
	_ context.Context, _ ...docker.NetworkListOption,
) (*client.NetworkListResult, error) {
	return &client.NetworkListResult{Items: f.networks}, nil
}

// fakeSettingRepo hands back one stored setting and one volume, and records what
// is written.
type fakeSettingRepo struct {
	repository.SettingRepo
	setting *entity.Setting
	// volume is what a lookup by id returns; nil means the volume is gone.
	volume   *entity.Setting
	upserted []*entity.Setting
}

func (f *fakeSettingRepo) GetSingle(
	_ context.Context, _ database.IDB, _ *entity.ObjectScope, _ base.SettingType,
	_ bool, _ ...bunex.SelectQueryOption,
) (*entity.Setting, error) {
	return f.setting, nil
}

func (f *fakeSettingRepo) GetByID(
	_ context.Context, _ database.IDB, _ *entity.ObjectScope, _ base.SettingType,
	_ string, _ bool, _ ...bunex.SelectQueryOption,
) (*entity.Setting, error) {
	if f.volume == nil {
		return nil, hperrors.Wrap(hperrors.ErrNotFound)
	}
	return f.volume, nil
}

func (f *fakeSettingRepo) Upsert(
	_ context.Context, _ database.IDB, setting *entity.Setting, _, _ []string, _ ...bunex.InsertQueryOption,
) error {
	f.upserted = append(f.upserted, setting)
	return nil
}

// sharedVolume is a volume shared with apps, which is what the backend needs.
func sharedVolume() *entity.Setting {
	raw, err := json.Marshal(&entity.ClusterVolume{})
	if err != nil {
		panic(err)
	}
	return &entity.Setting{
		ID: "vol-1", Name: "logs-data", Type: base.SettingTypeClusterVolume, Data: string(raw), Inheritable: true,
	}
}

// fakeSystemApps provisions apps the way systemappservice does as far as these
// tests can see: the document becomes the app's deployment settings, and
// Customize runs on the service it would be created with.
type fakeSystemApps struct {
	systemappservice.Service
	docker *fakeDocker

	apps        map[string]*entity.App
	provisioned []*systemappservice.ProvisionReq
	redeployed  []string
	secrets     map[string][]*systemappservice.SecretFile
	removed     []string
	// removedStorage says, by app key, whether its storage went with it.
	removedStorage map[string]bool
}

func (f *fakeSystemApps) LoadApp(_ context.Context, _ database.IDB, key string) (*entity.App, error) {
	return f.apps[key], nil
}

func (f *fakeSystemApps) Provision(
	_ context.Context, _ database.IDB, req *systemappservice.ProvisionReq,
) (*systemappservice.ProvisionResp, error) {
	f.provisioned = append(f.provisioned, req)

	spec := swarm.ServiceSpec{
		Mode: swarm.ServiceMode{Replicated: &swarm.ReplicatedService{Replicas: new(uint64(1))}},
		TaskTemplate: swarm.TaskSpec{
			ContainerSpec: &swarm.ContainerSpec{},
			Networks: []swarm.NetworkAttachmentConfig{
				{Target: "hivepaas_" + req.Env + "_net", Aliases: []string{req.Key}},
			},
		},
	}
	if req.Customize != nil {
		if err := req.Customize(&spec); err != nil {
			return nil, err
		}
	}
	serviceID := "svc-" + req.Key
	f.docker.inspected[serviceID] = swarm.Service{ID: serviceID, Spec: spec}

	source := req.Doc.Deployment.Source
	image, _ := source["imageSource"].(map[string]any)["image"].(string)
	command, _ := source["command"].(string)
	app := &entity.App{
		ID: "app-" + req.Key, Key: req.Key, Name: req.Name, ProjectID: "hivepaas", ServiceID: serviceID,
		Settings: []*entity.Setting{deploymentSetting(image, command)},
	}
	f.apps[req.Key] = app
	return &systemappservice.ProvisionResp{
		App:            app,
		DeploymentTask: &entity.Task{ID: "deploy-" + req.Key},
		Cleanup:        func(context.Context) error { return nil },
	}, nil
}

func (f *fakeSystemApps) Redeploy(
	_ context.Context, _ database.IDB, req *systemappservice.RedeployReq,
) (*entity.Task, error) {
	setting := req.App.GetSettingByType(base.SettingTypeAppDeployment)
	settings, err := setting.AsAppDeploymentSettings()
	if err != nil {
		return nil, err
	}
	if !req.Change(settings) {
		return nil, nil
	}
	if err = setting.SetData(settings); err != nil {
		return nil, err
	}
	f.redeployed = append(f.redeployed, req.App.Key)
	return &entity.Task{ID: "redeploy-" + req.App.Key}, nil
}

func (f *fakeSystemApps) SetResources(_ context.Context, app *entity.App, res systemappservice.Resources) error {
	svc := f.docker.inspected[app.ServiceID]
	systemappservice.ApplyResources(&svc.Spec, res)
	f.docker.inspected[app.ServiceID] = svc
	return nil
}

func (f *fakeSystemApps) SyncSecrets(
	_ context.Context, _ database.IDB, app *entity.App, files []*systemappservice.SecretFile,
) error {
	if f.secrets == nil {
		f.secrets = map[string][]*systemappservice.SecretFile{}
	}
	f.secrets[app.Key] = files
	return nil
}

func (f *fakeSystemApps) Remove(_ context.Context, _ database.IDB, app *entity.App, removeStorage bool) error {
	f.removed = append(f.removed, app.Key)
	if f.removedStorage == nil {
		f.removedStorage = map[string]bool{}
	}
	f.removedStorage[app.Key] = removeStorage
	delete(f.apps, app.Key)
	return nil
}

func deploymentSetting(image, command string) *entity.Setting {
	setting := &entity.Setting{ID: "deployment", Type: base.SettingTypeAppDeployment}
	setting.MustSetData(&entity.AppDeploymentSettings{
		ActiveMethod: base.DeploymentMethodImage,
		ImageSource:  &entity.DeploymentImageSource{Image: image},
		Command:      command,
	})
	return setting
}

func enabledConfig() *entity.LoggingSettings {
	return &entity.LoggingSettings{
		Enabled: true,
		Sources: entity.LoggingSources{Apps: true},
		Collector: entity.LoggingCollector{
			Type:    base.LoggingCollectorTypeVlagent,
			Managed: true,
		},
		Backend: entity.LoggingBackend{
			Type:    base.LoggingBackendTypeVictoriaLogs,
			Managed: true,
			VictoriaLogs: &entity.LoggingVictoriaLogs{
				Volume: entity.ObjectID{ID: "vol-1"},
			},
		},
	}
}

// storedSetting is shaped like a database row, with no parsed cache, so Apply
// reads what would really have persisted.
func storedSetting(t *testing.T, cfg *entity.LoggingSettings) *entity.Setting {
	t.Helper()
	// A credential is encrypted when stored, which takes a key.
	key, err := datakey.Generate()
	if err != nil {
		t.Fatalf("datakey.Generate: %v", err)
	}
	datakey.SetActive(key)

	s := &entity.Setting{ID: "s1", Type: base.SettingTypeLogging}
	if err := s.SetData(cfg); err != nil {
		t.Fatalf("SetData: %v", err)
	}
	return &entity.Setting{ID: s.ID, Type: s.Type, Data: s.Data}
}

// newTestService builds the service with every dependency faked. setting may be
// nil for "never configured".
func newTestService(fd *fakeDocker, setting *entity.Setting) *service {
	if fd.inspected == nil {
		fd.inspected = map[string]swarm.Service{}
	}
	// The cluster HivePaaS runs in always has this network; a test for its
	// absence removes it again.
	fd.networks = append(fd.networks, network.Summary{Network: network.Network{
		ID: "id-local", Name: base.NetworkHivepaasLocal,
	}})
	return &service{
		dockerManager:    fd,
		settingRepo:      &fakeSettingRepo{setting: setting, volume: sharedVolume()},
		systemAppService: &fakeSystemApps{docker: fd, apps: map[string]*entity.App{}},
		logger:           logging.GlobalLogger(),
	}
}

func appsOf(s *service) *fakeSystemApps {
	return s.systemAppService.(*fakeSystemApps) //nolint:forcetypeassert // tests build it
}

func settingsOf(s *service) *fakeSettingRepo {
	return s.settingRepo.(*fakeSettingRepo) //nolint:forcetypeassert // tests build it
}
