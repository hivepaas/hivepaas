package loggingserviceimpl

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/moby/moby/api/types/network"
	"github.com/moby/moby/api/types/swarm"
	"github.com/moby/moby/client"
	"github.com/stretchr/testify/assert"

	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/basedto"
	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/infra/database"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/bunex"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/unit"
	"github.com/hivepaas/hivepaas/hivepaas_app/repository"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/loggingservice"
	"github.com/hivepaas/hivepaas/services/docker"
)

// fakeDocker records what was asked of it. Embedding the interface means only
// the methods this path uses need a body; anything else panics loudly instead
// of passing quietly.
type fakeDocker struct {
	docker.Manager
	existing map[string]bool
	listed   []swarm.Service
	created  []*swarm.ServiceSpec
	updated  []*swarm.ServiceSpec
	removed  []string

	networks        []network.Summary
	networksCreated []client.NetworkCreateOptions
	inspected       map[string]swarm.Service
}

func (f *fakeDocker) ServiceList(
	_ context.Context, _ ...docker.ServiceListOption,
) (*client.ServiceListResult, error) {
	return &client.ServiceListResult{Items: f.listed}, nil
}

// fakeAppRepo hands back a fixed set of apps.
type fakeAppRepo struct {
	repository.AppRepo
	apps []*entity.App
}

func (f *fakeAppRepo) List(
	_ context.Context, _ database.IDB, _ string, _ *basedto.Paging, _ ...bunex.SelectQueryOption,
) ([]*entity.App, *basedto.PagingMeta, error) {
	return f.apps, &basedto.PagingMeta{}, nil
}

func (f *fakeDocker) ServiceInspect(
	_ context.Context, serviceID string, _ ...docker.ServiceInspectOption,
) (*client.ServiceInspectResult, error) {
	if svc, ok := f.inspected[serviceID]; ok {
		return &client.ServiceInspectResult{Service: svc}, nil
	}
	if !f.existing[serviceID] {
		return nil, hperrors.Wrap(hperrors.ErrNotFound)
	}
	return &client.ServiceInspectResult{Service: swarm.Service{ID: "svc-" + serviceID}}, nil
}

func (f *fakeDocker) ServiceCreate(
	_ context.Context, spec *swarm.ServiceSpec, _ ...docker.ServiceCreateOption,
) (*client.ServiceCreateResult, error) {
	if f.existing[spec.Name] {
		return nil, hperrors.Wrap(hperrors.ErrConflict) // what docker says to a name clash
	}
	if f.existing == nil {
		f.existing = map[string]bool{}
	}
	f.existing[spec.Name] = true
	f.created = append(f.created, spec)
	return &client.ServiceCreateResult{ID: "svc-" + spec.Name}, nil
}

func (f *fakeDocker) ServiceUpdate(
	_ context.Context, _ string, _ *swarm.Version, spec *swarm.ServiceSpec, _ ...docker.ServiceUpdateOption,
) (*client.ServiceUpdateResult, error) {
	f.updated = append(f.updated, spec)
	return &client.ServiceUpdateResult{}, nil
}

func (f *fakeDocker) ServiceRemove(
	_ context.Context, serviceID string, _ ...docker.ServiceRemoveOption,
) (*client.ServiceRemoveResult, error) {
	f.removed = append(f.removed, serviceID)
	if !f.existing[serviceID] {
		return nil, hperrors.Wrap(hperrors.ErrNotFound)
	}
	delete(f.existing, serviceID)
	return &client.ServiceRemoveResult{}, nil
}

// fakeSettingRepo hands back one stored setting, or none.
type fakeSettingRepo struct {
	repository.SettingRepo
	setting *entity.Setting
	// volume is what a lookup by id returns; nil means the volume is pinned to
	// nowhere, which is a placement answer rather than a failure.
	volume *entity.Setting
}

func (f *fakeSettingRepo) GetByID(
	_ context.Context, _ database.IDB, _ *entity.ObjectScope, _ base.SettingType,
	_ string, _ bool, _ ...bunex.SelectQueryOption,
) (*entity.Setting, error) {
	if f.volume != nil {
		return f.volume, nil
	}
	return unpinnedVolumeSetting(), nil
}

// unpinnedVolumeSetting is a volume that names no node, so placement is free.
func unpinnedVolumeSetting() *entity.Setting {
	return volumeSetting(&entity.ClusterVolume{})
}

func volumeSetting(vol *entity.ClusterVolume) *entity.Setting {
	raw, err := json.Marshal(vol)
	if err != nil {
		panic(err)
	}
	return &entity.Setting{
		ID: "vol-1", Name: "logs-data", Type: base.SettingTypeClusterVolume, Data: string(raw),
	}
}

func (f *fakeSettingRepo) GetSingle(
	_ context.Context, _ database.IDB, _ *entity.ObjectScope, _ base.SettingType,
	_ bool, _ ...bunex.SelectQueryOption,
) (*entity.Setting, error) {
	return f.setting, nil
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
	s := &entity.Setting{ID: "s1", Type: base.SettingTypeLogging}
	if err := s.SetData(cfg); err != nil {
		t.Fatalf("SetData: %v", err)
	}
	return &entity.Setting{ID: s.ID, Type: s.Type, Data: s.Data}
}

// newTestService builds the service with every dependency faked. setting may be
// nil for "never configured".
func newTestService(fd *fakeDocker, setting *entity.Setting) *service {
	// The cluster HivePaaS runs in always has this network; a test for its
	// absence removes it again.
	fd.networks = append(fd.networks, network.Summary{Network: network.Network{
		ID: "id-local", Name: base.NetworkHivepaasLocal,
	}})
	return &service{
		dockerManager: fd,
		settingRepo:   &fakeSettingRepo{setting: setting},
		appRepo:       &fakeAppRepo{},
	}
}

func TestDeployCreatesBackendBeforeCollector(t *testing.T) {
	fd := &fakeDocker{}
	s := newTestService(fd, nil)

	if err := s.deploy(context.Background(), nil, enabledConfig()); err != nil {
		t.Fatalf("deploy: %v", err)
	}

	if len(fd.created) != 2 {
		t.Fatalf("want two services, got %d", len(fd.created))
	}
	assert.Equal(t, ServiceNameBackend, fd.created[0].Name,
		"the collector must not be started before something to ship to exists")
	assert.Equal(t, ServiceNameCollector, fd.created[1].Name)
}

func TestDeployRunsTheCollectorOnEveryNode(t *testing.T) {
	fd := &fakeDocker{}
	s := newTestService(fd, nil)

	if err := s.deploy(context.Background(), nil, enabledConfig()); err != nil {
		t.Fatalf("deploy: %v", err)
	}

	collector := fd.created[1]
	assert.NotNil(t, collector.Mode.Global)
	assert.Nil(t, collector.Mode.Replicated)
}

// The collector must write to the backend it was deployed beside, by the
// service name the overlay network resolves.
func TestDeployPointsTheCollectorAtTheBackend(t *testing.T) {
	fd := &fakeDocker{}
	s := newTestService(fd, nil)

	if err := s.deploy(context.Background(), nil, enabledConfig()); err != nil {
		t.Fatalf("deploy: %v", err)
	}

	assert.Contains(t, fd.created[1].TaskTemplate.ContainerSpec.Args,
		"-remoteWrite.url=http://"+ServiceNameBackend+":9428/internal/insert")
}

func TestDeployRefusesAManagedBackendWithNoVolume(t *testing.T) {
	cfg := enabledConfig()
	cfg.Backend.VictoriaLogs.Volume.ID = ""
	fd := &fakeDocker{}
	s := newTestService(fd, nil)

	err := s.deploy(context.Background(), nil, cfg)

	// The service layer refuses first, with its own code. The backend package
	// refuses too, but reaching it would mean the check here had gone missing.
	assert.ErrorIs(t, err, hperrors.ErrLoggingVolumeMissing, "a backend with no volume loses every log on restart")
	assert.Empty(t, fd.created, "nothing should be half-deployed")
}

// Placement follows the volume, so the two can never name different nodes.
func TestDeployPinsTheBackendWhereItsVolumeIs(t *testing.T) {
	fd := &fakeDocker{}
	s := newTestService(fd, nil)
	s.settingRepo = &fakeSettingRepo{volume: volumeSetting(&entity.ClusterVolume{NodeID: "node-7"})}

	if err := s.deploy(context.Background(), nil, enabledConfig()); err != nil {
		t.Fatalf("deploy: %v", err)
	}
	backend := fd.created[0]
	if assert.NotNil(t, backend.TaskTemplate.Placement) {
		assert.Equal(t, []string{"node.id==node-7"}, backend.TaskTemplate.Placement.Constraints)
	}
}

// A volume pinned by label pins the backend the same way - something the node
// field this replaced could not express at all.
func TestDeployPinsTheBackendByVolumeNodeLabel(t *testing.T) {
	fd := &fakeDocker{}
	s := newTestService(fd, nil)
	s.settingRepo = &fakeSettingRepo{volume: volumeSetting(&entity.ClusterVolume{NodeLabel: "storage=fast"})}

	if err := s.deploy(context.Background(), nil, enabledConfig()); err != nil {
		t.Fatalf("deploy: %v", err)
	}
	backend := fd.created[0]
	if assert.NotNil(t, backend.TaskTemplate.Placement) {
		assert.Equal(t, []string{"node.labels.storage==fast"}, backend.TaskTemplate.Placement.Constraints)
	}
}

// An unpinned volume leaves the backend unpinned: the volume decides placement
// completely, including deciding not to.
func TestDeployLeavesTheBackendUnpinnedForAnUnpinnedVolume(t *testing.T) {
	fd := &fakeDocker{}
	s := newTestService(fd, nil)

	if err := s.deploy(context.Background(), nil, enabledConfig()); err != nil {
		t.Fatalf("deploy: %v", err)
	}
	assert.Nil(t, fd.created[0].TaskTemplate.Placement)
}

func TestDeployAppliesBackendResourceLimits(t *testing.T) {
	fd := &fakeDocker{}
	s := newTestService(fd, nil)
	cfg := enabledConfig()
	cfg.Backend.VictoriaLogs.CPULimit = 2
	cfg.Backend.VictoriaLogs.MemoryLimit = unit.MustParseDataSizeString("1gb")

	if err := s.deploy(context.Background(), nil, cfg); err != nil {
		t.Fatalf("deploy: %v", err)
	}
	backend := fd.created[0]
	if assert.NotNil(t, backend.TaskTemplate.Resources) {
		assert.Equal(t, int64(2*docker.UnitCPUNano), backend.TaskTemplate.Resources.Limits.NanoCPUs)
		assert.Equal(t, int64(1<<30), backend.TaskTemplate.Resources.Limits.MemoryBytes)
		assert.Nil(t, backend.TaskTemplate.Resources.Reservations,
			"a reservation could leave the backend unschedulable")
	}
}

// An unmanaged backend is somebody else's service: HivePaaS ships to it and
// never creates it.
func TestDeploySkipsAnUnmanagedBackend(t *testing.T) {
	cfg := enabledConfig()
	cfg.Backend.Managed = false
	cfg.Backend.VictoriaLogs = nil
	cfg.Backend.Ingest = &entity.LoggingEndpoint{URL: "https://logs.example/insert"}
	fd := &fakeDocker{}
	s := newTestService(fd, nil)

	if err := s.deploy(context.Background(), nil, cfg); err != nil {
		t.Fatalf("deploy: %v", err)
	}

	if len(fd.created) != 1 {
		t.Fatalf("want only the collector, got %d services", len(fd.created))
	}
	assert.Equal(t, ServiceNameCollector, fd.created[0].Name)
	assert.Contains(t, fd.created[0].TaskTemplate.ContainerSpec.Args,
		"-remoteWrite.url=https://logs.example/insert")
}

func TestTearDownRemovesCollectorBeforeBackend(t *testing.T) {
	fd := &fakeDocker{existing: map[string]bool{ServiceNameCollector: true, ServiceNameBackend: true}}
	s := newTestService(fd, nil)

	if err := s.TearDown(context.Background()); err != nil {
		t.Fatalf("TearDown: %v", err)
	}

	assert.Equal(t, []string{ServiceNameCollector, ServiceNameBackend}, fd.removed,
		"stop shipping before removing the thing being shipped to")
}

// Turning logging off must not destroy the logs. For a plain local volume the
// data is inside the volume, so removing it is unrecoverable; removal is a
// separate, explicit action through the volume API.
func TestTearDownKeepsTheDataVolume(t *testing.T) {
	fd := &fakeDocker{existing: map[string]bool{ServiceNameCollector: true, ServiceNameBackend: true}}
	s := newTestService(fd, nil)

	if err := s.TearDown(context.Background()); err != nil {
		t.Fatalf("TearDown: %v", err)
	}

	for _, r := range fd.removed {
		assert.NotContains(t, r, "vol-", "the data volume must survive tear-down")
	}
}

// Never configured is the default state, and must deploy nothing.
func TestApplyWithNoSettingDoesNothing(t *testing.T) {
	fd := &fakeDocker{}
	s := newTestService(fd, nil)

	_, e := s.Apply(context.Background(), nil, &loggingservice.SettingApplyReq{})
	assert.NoError(t, e)
	assert.Empty(t, fd.created)
	assert.Empty(t, fd.removed)
}

func TestApplyWhenDisabledTearsDown(t *testing.T) {
	cfg := enabledConfig()
	cfg.Enabled = false
	fd := &fakeDocker{existing: map[string]bool{ServiceNameCollector: true, ServiceNameBackend: true}}
	s := newTestService(fd, storedSetting(t, cfg))

	_, e := s.Apply(context.Background(), nil, &loggingservice.SettingApplyReq{})
	assert.NoError(t, e)
	assert.Empty(t, fd.created)
	assert.Equal(t, []string{ServiceNameCollector, ServiceNameBackend}, fd.removed)
}

func TestApplyWhenEnabledDeploys(t *testing.T) {
	fd := &fakeDocker{}
	s := newTestService(fd, storedSetting(t, enabledConfig()))

	_, e := s.Apply(context.Background(), nil, &loggingservice.SettingApplyReq{})
	assert.NoError(t, e)
	assert.Len(t, fd.created, 2)
}

// Apply runs on every save. A second run must update what the first created,
// not try to create it again and fail on the name.
func TestDeployTwiceUpdatesInsteadOfCreating(t *testing.T) {
	fd := &fakeDocker{}
	s := newTestService(fd, nil)

	if err := s.deploy(context.Background(), nil, enabledConfig()); err != nil {
		t.Fatalf("first deploy: %v", err)
	}
	if err := s.deploy(context.Background(), nil, enabledConfig()); err != nil {
		t.Fatalf("second deploy: %v", err)
	}

	assert.Len(t, fd.created, 2, "each service created once")
	assert.Len(t, fd.updated, 2, "and updated on the second run")
}

// Switching to the operator's own collector must stop HivePaaS's, or both ship.
func TestDeployRemovesTheCollectorWhenItStopsBeingManaged(t *testing.T) {
	fd := &fakeDocker{}
	s := newTestService(fd, nil)
	if err := s.deploy(context.Background(), nil, enabledConfig()); err != nil {
		t.Fatalf("deploy: %v", err)
	}

	cfg := enabledConfig()
	cfg.Collector.Managed = false
	if err := s.deploy(context.Background(), nil, cfg); err != nil {
		t.Fatalf("redeploy: %v", err)
	}

	assert.Equal(t, []string{ServiceNameCollector}, fd.removed)
}

// Moving to a backend HivePaaS does not run removes the one it did - after the
// collector has been repointed, and without touching the data volume.
func TestDeployRemovesTheBackendWhenItStopsBeingManaged(t *testing.T) {
	fd := &fakeDocker{}
	s := newTestService(fd, nil)
	if err := s.deploy(context.Background(), nil, enabledConfig()); err != nil {
		t.Fatalf("deploy: %v", err)
	}

	cfg := enabledConfig()
	cfg.Backend.Managed = false
	cfg.Backend.VictoriaLogs = nil
	cfg.Backend.Ingest = &entity.LoggingEndpoint{URL: "https://logs.example/insert"}
	if err := s.deploy(context.Background(), nil, cfg); err != nil {
		t.Fatalf("redeploy: %v", err)
	}

	assert.Equal(t, []string{ServiceNameBackend}, fd.removed)
	last := fd.updated[len(fd.updated)-1]
	assert.Equal(t, ServiceNameCollector, last.Name, "the collector is repointed before the backend goes")
	assert.Contains(t, last.TaskTemplate.ContainerSpec.Args, "-remoteWrite.url=https://logs.example/insert")
}
