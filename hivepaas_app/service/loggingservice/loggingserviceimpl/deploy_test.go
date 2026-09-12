package loggingserviceimpl

import (
	"context"
	"testing"

	"github.com/moby/moby/api/types/swarm"
	"github.com/moby/moby/client"
	"github.com/stretchr/testify/assert"

	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/infra/database"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/bunex"
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
	created  []*swarm.ServiceSpec
	updated  []*swarm.ServiceSpec
	removed  []string
}

func (f *fakeDocker) ServiceInspect(
	_ context.Context, serviceID string, _ ...docker.ServiceInspectOption,
) (*client.ServiceInspectResult, error) {
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
}

func (f *fakeSettingRepo) GetSingle(
	_ context.Context, _ database.IDB, _ *entity.ObjectScope, _ base.SettingType,
	_ bool, _ ...bunex.SelectQueryOption,
) (*entity.Setting, error) {
	return f.setting, nil
}

func enabledConfig() *entity.Logging {
	return &entity.Logging{
		Enabled: true,
		Sources: entity.LoggingSources{Apps: true},
		Collector: entity.LoggingCollector{
			Type:    entity.LoggingCollectorTypeVlagent,
			Managed: true,
		},
		Backend: entity.LoggingBackend{
			Type:    entity.LoggingBackendTypeVictoriaLogs,
			Managed: true,
			VictoriaLogs: &entity.LoggingVictoriaLogs{
				NodeID:   "node-1",
				VolumeID: "vol-1",
			},
		},
	}
}

// storedSetting is shaped like a database row, with no parsed cache, so Apply
// reads what would really have persisted.
func storedSetting(t *testing.T, cfg *entity.Logging) *entity.Setting {
	t.Helper()
	s := &entity.Setting{ID: "s1", Type: base.SettingTypeLogging}
	if err := s.SetData(cfg); err != nil {
		t.Fatalf("SetData: %v", err)
	}
	return &entity.Setting{ID: s.ID, Type: s.Type, Data: s.Data}
}

func TestDeployCreatesBackendBeforeCollector(t *testing.T) {
	fd := &fakeDocker{}
	s := &service{dockerManager: fd}

	if err := s.deploy(context.Background(), enabledConfig()); err != nil {
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
	s := &service{dockerManager: fd}

	if err := s.deploy(context.Background(), enabledConfig()); err != nil {
		t.Fatalf("deploy: %v", err)
	}

	collector := fd.created[1]
	assert.NotNil(t, collector.Mode.Global)
	assert.Nil(t, collector.Mode.Replicated)
}

func TestDeployPinsTheBackendToItsNode(t *testing.T) {
	fd := &fakeDocker{}
	s := &service{dockerManager: fd}

	if err := s.deploy(context.Background(), enabledConfig()); err != nil {
		t.Fatalf("deploy: %v", err)
	}

	backend := fd.created[0]
	if backend.TaskTemplate.Placement == nil {
		t.Fatal("the backend is not pinned")
	}
	assert.Contains(t, backend.TaskTemplate.Placement.Constraints, "node.id==node-1")
}

// The collector must write to the backend it was deployed beside, by the
// service name the overlay network resolves.
func TestDeployPointsTheCollectorAtTheBackend(t *testing.T) {
	fd := &fakeDocker{}
	s := &service{dockerManager: fd}

	if err := s.deploy(context.Background(), enabledConfig()); err != nil {
		t.Fatalf("deploy: %v", err)
	}

	assert.Contains(t, fd.created[1].TaskTemplate.ContainerSpec.Args,
		"-remoteWrite.url=http://"+ServiceNameBackend+":9428/internal/insert")
}

func TestDeployRefusesAManagedBackendWithNoVolume(t *testing.T) {
	cfg := enabledConfig()
	cfg.Backend.VictoriaLogs.VolumeID = ""
	fd := &fakeDocker{}
	s := &service{dockerManager: fd}

	err := s.deploy(context.Background(), cfg)

	// The service layer refuses first, with its own code. The backend package
	// refuses too, but reaching it would mean the check here had gone missing.
	assert.ErrorIs(t, err, loggingservice.ErrVolumeMissing, "a backend with no volume loses every log on restart")
	assert.Empty(t, fd.created, "nothing should be half-deployed")
}

func TestDeployRefusesAManagedBackendWithNoNode(t *testing.T) {
	cfg := enabledConfig()
	cfg.Backend.VictoriaLogs.NodeID = ""
	s := &service{dockerManager: &fakeDocker{}}

	assert.ErrorIs(t, s.deploy(context.Background(), cfg), loggingservice.ErrBackendNodeMissing)
}

// An unmanaged backend is somebody else's service: HivePaaS ships to it and
// never creates it.
func TestDeploySkipsAnUnmanagedBackend(t *testing.T) {
	cfg := enabledConfig()
	cfg.Backend.Managed = false
	cfg.Backend.VictoriaLogs = nil
	cfg.Backend.Ingest = &entity.LoggingEndpoint{URL: "https://logs.example/insert"}
	fd := &fakeDocker{}
	s := &service{dockerManager: fd}

	if err := s.deploy(context.Background(), cfg); err != nil {
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
	s := &service{dockerManager: fd}

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
	s := &service{dockerManager: fd}

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
	s := &service{dockerManager: fd, settingRepo: &fakeSettingRepo{}}

	assert.NoError(t, s.Apply(context.Background(), nil))
	assert.Empty(t, fd.created)
	assert.Empty(t, fd.removed)
}

func TestApplyWhenDisabledTearsDown(t *testing.T) {
	cfg := enabledConfig()
	cfg.Enabled = false
	fd := &fakeDocker{existing: map[string]bool{ServiceNameCollector: true, ServiceNameBackend: true}}
	s := &service{dockerManager: fd, settingRepo: &fakeSettingRepo{setting: storedSetting(t, cfg)}}

	assert.NoError(t, s.Apply(context.Background(), nil))
	assert.Empty(t, fd.created)
	assert.Equal(t, []string{ServiceNameCollector, ServiceNameBackend}, fd.removed)
}

func TestApplyWhenEnabledDeploys(t *testing.T) {
	fd := &fakeDocker{}
	s := &service{dockerManager: fd, settingRepo: &fakeSettingRepo{setting: storedSetting(t, enabledConfig())}}

	assert.NoError(t, s.Apply(context.Background(), nil))
	assert.Len(t, fd.created, 2)
}

// Apply runs on every save. A second run must update what the first created,
// not try to create it again and fail on the name.
func TestDeployTwiceUpdatesInsteadOfCreating(t *testing.T) {
	fd := &fakeDocker{}
	s := &service{dockerManager: fd}

	if err := s.deploy(context.Background(), enabledConfig()); err != nil {
		t.Fatalf("first deploy: %v", err)
	}
	if err := s.deploy(context.Background(), enabledConfig()); err != nil {
		t.Fatalf("second deploy: %v", err)
	}

	assert.Len(t, fd.created, 2, "each service created once")
	assert.Len(t, fd.updated, 2, "and updated on the second run")
}

// Switching to the operator's own collector must stop HivePaaS's, or both ship.
func TestDeployRemovesTheCollectorWhenItStopsBeingManaged(t *testing.T) {
	fd := &fakeDocker{}
	s := &service{dockerManager: fd}
	if err := s.deploy(context.Background(), enabledConfig()); err != nil {
		t.Fatalf("deploy: %v", err)
	}

	cfg := enabledConfig()
	cfg.Collector.Managed = false
	if err := s.deploy(context.Background(), cfg); err != nil {
		t.Fatalf("redeploy: %v", err)
	}

	assert.Equal(t, []string{ServiceNameCollector}, fd.removed)
}

// Moving to a backend HivePaaS does not run removes the one it did - after the
// collector has been repointed, and without touching the data volume.
func TestDeployRemovesTheBackendWhenItStopsBeingManaged(t *testing.T) {
	fd := &fakeDocker{}
	s := &service{dockerManager: fd}
	if err := s.deploy(context.Background(), enabledConfig()); err != nil {
		t.Fatalf("deploy: %v", err)
	}

	cfg := enabledConfig()
	cfg.Backend.Managed = false
	cfg.Backend.VictoriaLogs = nil
	cfg.Backend.Ingest = &entity.LoggingEndpoint{URL: "https://logs.example/insert"}
	if err := s.deploy(context.Background(), cfg); err != nil {
		t.Fatalf("redeploy: %v", err)
	}

	assert.Equal(t, []string{ServiceNameBackend}, fd.removed)
	last := fd.updated[len(fd.updated)-1]
	assert.Equal(t, ServiceNameCollector, last.Name, "the collector is repointed before the backend goes")
	assert.Contains(t, last.TaskTemplate.ContainerSpec.Args, "-remoteWrite.url=https://logs.example/insert")
}
