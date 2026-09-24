package sysupdateserviceimpl

import (
	"context"
	"sync"
	"testing"
	"time"

	"github.com/moby/moby/api/types/swarm"
	"github.com/moby/moby/client"
	"github.com/stretchr/testify/assert"

	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/infra/database"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/tasklog"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/systemappservice"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/sysupdateservice"
	"github.com/hivepaas/hivepaas/hivepaas_app/tasks/queue"
	"github.com/hivepaas/hivepaas/services/docker"
)

// fakeDocker answers for the two logging services and records what it was asked
// to update. Embedding the interface means only the methods this path uses need
// a body; anything else panics loudly rather than passing quietly.
type fakeDocker struct {
	docker.Manager
	// The app and the worker are updated concurrently, so the recording has to
	// survive two goroutines. The code under test shares nothing between them;
	// this bookkeeping does.
	mu sync.Mutex

	deployed map[string]string // service name -> running image
	updated  map[string]string // service name -> image it was updated to
	specs    map[string]*swarm.ServiceSpec

	// replicaUpdates records every replica count written, with the state of the
	// context at the moment of the call.
	replicaUpdates []replicaUpdate
	ctxErr         error

	// onUpdate, when set, is told about every image an update writes - a test
	// that reads the service again afterwards uses it to make the update land.
	onUpdate func(serviceID, image string)
}

type replicaUpdate struct {
	replicas uint64
	ctxErr   error
}

func (f *fakeDocker) ServiceInspect(
	_ context.Context, serviceID string, _ ...docker.ServiceInspectOption,
) (*client.ServiceInspectResult, error) {
	image, ok := f.deployed[serviceID]
	if !ok {
		return nil, hperrors.Wrap(hperrors.ErrInfraNotFound)
	}
	return &client.ServiceInspectResult{Service: swarm.Service{
		ID: serviceID,
		Spec: swarm.ServiceSpec{
			Annotations:  swarm.Annotations{Name: serviceID},
			TaskTemplate: swarm.TaskSpec{ContainerSpec: &swarm.ContainerSpec{Image: image}},
		},
	}}, nil
}

func (f *fakeDocker) ServiceUpdate(
	ctx context.Context, serviceID string, _ *swarm.Version, spec *swarm.ServiceSpec,
	_ ...docker.ServiceUpdateOption,
) (*client.ServiceUpdateResult, error) {
	f.ctxErr = ctx.Err()
	return &client.ServiceUpdateResult{}, f.record(serviceID, spec)
}

func (f *fakeDocker) record(serviceID string, spec *swarm.ServiceSpec) error {
	f.mu.Lock()
	defer f.mu.Unlock()

	if spec.Mode.Replicated != nil && spec.Mode.Replicated.Replicas != nil {
		f.replicaUpdates = append(f.replicaUpdates, replicaUpdate{
			replicas: *spec.Mode.Replicated.Replicas,
			ctxErr:   f.ctxErr,
		})
	}
	if f.updated == nil {
		f.updated = map[string]string{}
		f.specs = map[string]*swarm.ServiceSpec{}
	}
	if spec.TaskTemplate.ContainerSpec != nil {
		f.updated[serviceID] = spec.TaskTemplate.ContainerSpec.Image
		if f.onUpdate != nil {
			f.onUpdate(serviceID, spec.TaskTemplate.ContainerSpec.Image)
		}
	}
	f.specs[serviceID] = spec
	return nil
}

// ServiceUpdateFunc mirrors what the manager does: hand the callback the service,
// respect its "nothing to do" answer, and only then record an update. Retries are
// not simulated - nothing here fails - so one attempt is the whole behavior.
func (f *fakeDocker) ServiceUpdateFunc(
	ctx context.Context, serviceID string, service *swarm.Service,
	fn func(int, *swarm.Service) (bool, error), _ int, _ time.Duration,
	_ ...docker.ServiceUpdateOption,
) error {
	f.mu.Lock()
	f.ctxErr = ctx.Err()
	f.mu.Unlock()

	apply, err := fn(0, service)
	if err != nil || !apply {
		return err
	}
	return f.record(serviceID, &service.Spec)
}

func (f *fakeDocker) ServiceUpdateWait(
	_ context.Context, serviceID string, _ time.Duration,
) (*swarm.Service, error) {
	return &swarm.Service{ID: serviceID}, nil
}

func loggingUpdateData(t *testing.T, target *base.ReleaseInfo) *sysUpdateData {
	t.Helper()

	task := &entity.Task{ID: "t1", Type: base.TaskTypeSystemUpdate}
	task.MustSetArgs(&entity.TaskSystemUpdateArgs{
		CurrentVersion: base.StableVersion,
		TargetVersion:  target,
	})
	return &sysUpdateData{
		SysUpdateReq: &sysupdateservice.SysUpdateReq{
			TaskExecData: &queue.TaskExecData{Task: task, LogStore: tasklog.NewLocalStore("t1")},
		},
	}
}

// fakeSystemApps answers for the logging stack's apps and records the images
// written into their deployment settings.
type fakeSystemApps struct {
	systemappservice.Service
	apps     map[string]*entity.App
	recorded map[string]string // app key -> image recorded
}

func (f *fakeSystemApps) LoadApp(_ context.Context, _ database.IDB, key string) (*entity.App, error) {
	return f.apps[key], nil
}

func (f *fakeSystemApps) RecordImage(_ context.Context, _ database.IDB, app *entity.App, image string) error {
	if f.recorded == nil {
		f.recorded = map[string]string{}
	}
	f.recorded[app.Key] = image
	return nil
}

// loggingStack deploys the logging apps named in running, each on the image
// given, the way provisioning leaves them: an app with a service of its own.
func loggingStack(running map[string]string) (*fakeDocker, *fakeSystemApps) {
	f := &fakeDocker{deployed: map[string]string{}}
	apps := &fakeSystemApps{apps: map[string]*entity.App{}}
	for key, image := range running {
		serviceID := "svc-" + key
		apps.apps[key] = &entity.App{ID: "app-" + key, Key: key, ServiceID: serviceID}
		f.deployed[serviceID] = image
	}
	return f, apps
}

func TestLoggingUpdateMovesBothAppsToANewerImage(t *testing.T) {
	f, apps := loggingStack(map[string]string{
		base.HivepaasVictoriaLogsKey: "victoriametrics/victoria-logs:v1.52.0",
		base.HivepaasVlagentKey:      "victoriametrics/vlagent:v1.52.0",
	})
	s := &service{dockerManager: f, systemAppService: apps}

	err := s.updateLoggingService(context.Background(), nil, loggingUpdateData(t, &base.ReleaseInfo{
		VictoriaLogsImage: "victoriametrics/victoria-logs:v1.53.0",
		VlagentImage:      "victoriametrics/vlagent:v1.53.0",
	}))

	assert.NoError(t, err)
	assert.Equal(t, map[string]string{
		"svc-victoria-logs": "victoriametrics/victoria-logs:v1.53.0",
		"svc-vlagent":       "victoriametrics/vlagent:v1.53.0",
	}, f.updated)
}

// The service moves with a monitored rollback, and the app's deployment settings
// then say so: otherwise the app's next deployment would put the old image back.
func TestLoggingUpdateRecordsTheImageInTheAppsSettings(t *testing.T) {
	f, apps := loggingStack(map[string]string{
		base.HivepaasVictoriaLogsKey: "victoriametrics/victoria-logs:v1.52.0",
	})
	f.onUpdate = func(serviceID, image string) { f.deployed[serviceID] = image }
	s := &service{dockerManager: f, systemAppService: apps}

	err := s.updateLoggingService(context.Background(), nil, loggingUpdateData(t, &base.ReleaseInfo{
		VictoriaLogsImage: "victoriametrics/victoria-logs:v1.53.0",
	}))

	assert.NoError(t, err)
	assert.Equal(t, map[string]string{
		base.HivepaasVictoriaLogsKey: "victoriametrics/victoria-logs:v1.53.0",
	}, apps.recorded)
}

// Re-running the same release must not restart the logging stack. This is the
// whole point of the version check.
func TestLoggingUpdateLeavesTheSameVersionAlone(t *testing.T) {
	f, apps := loggingStack(map[string]string{
		base.HivepaasVictoriaLogsKey: "victoriametrics/victoria-logs:v1.52.0",
		base.HivepaasVlagentKey:      "victoriametrics/vlagent:v1.52.0",
	})
	s := &service{dockerManager: f, systemAppService: apps}

	err := s.updateLoggingService(context.Background(), nil, loggingUpdateData(t, &base.ReleaseInfo{
		VictoriaLogsImage: "victoriametrics/victoria-logs:v1.52.0",
		VlagentImage:      "victoriametrics/vlagent:v1.52.0",
	}))

	assert.NoError(t, err)
	assert.Empty(t, f.updated)
}

// A digest is what a running service actually reports, and it moves whenever a
// tag is re-pushed. Same tag is still the same version - and what the settings
// record is the release's name for it, not the digest.
func TestLoggingUpdateIgnoresADigestUnderTheSameTag(t *testing.T) {
	f, apps := loggingStack(map[string]string{
		base.HivepaasVictoriaLogsKey: "victoriametrics/victoria-logs:v1.52.0@sha256:aaaa",
	})
	s := &service{dockerManager: f, systemAppService: apps}

	err := s.updateLoggingService(context.Background(), nil, loggingUpdateData(t, &base.ReleaseInfo{
		VictoriaLogsImage: "victoriametrics/victoria-logs:v1.52.0",
	}))

	assert.NoError(t, err)
	assert.Empty(t, f.updated)
	assert.Equal(t, "victoriametrics/victoria-logs:v1.52.0", apps.recorded[base.HivepaasVictoriaLogsKey])
}

// Nothing moves backwards, and the settings keep saying what runs.
func TestLoggingUpdateRefusesToGoBackwards(t *testing.T) {
	f, apps := loggingStack(map[string]string{
		base.HivepaasVictoriaLogsKey: "victoriametrics/victoria-logs:v1.52.0",
	})
	s := &service{dockerManager: f, systemAppService: apps}

	err := s.updateLoggingService(context.Background(), nil, loggingUpdateData(t, &base.ReleaseInfo{
		VictoriaLogsImage: "victoriametrics/victoria-logs:v1.51.0",
	}))

	assert.NoError(t, err)
	assert.Empty(t, f.updated)
	assert.Empty(t, apps.recorded)
}

// Logging is optional. Neither app existing is the ordinary state of an install
// that never switched it on, and it must not fail the whole update.
func TestLoggingUpdateSkipsWhatIsNotDeployed(t *testing.T) {
	f, apps := loggingStack(nil)
	s := &service{dockerManager: f, systemAppService: apps}

	err := s.updateLoggingService(context.Background(), nil, loggingUpdateData(t, &base.ReleaseInfo{
		VictoriaLogsImage: "victoriametrics/victoria-logs:v1.53.0",
		VlagentImage:      "victoriametrics/vlagent:v1.53.0",
	}))

	assert.NoError(t, err)
	assert.Empty(t, f.updated)
	assert.Empty(t, apps.recorded)
}

// The collector runs but the backend is somebody else's: only the app that is
// there moves.
func TestLoggingUpdateMovesOnlyTheAppThatExists(t *testing.T) {
	f, apps := loggingStack(map[string]string{
		base.HivepaasVlagentKey: "victoriametrics/vlagent:v1.52.0",
	})
	s := &service{dockerManager: f, systemAppService: apps}

	err := s.updateLoggingService(context.Background(), nil, loggingUpdateData(t, &base.ReleaseInfo{
		VictoriaLogsImage: "victoriametrics/victoria-logs:v1.53.0",
		VlagentImage:      "victoriametrics/vlagent:v1.53.0",
	}))

	assert.NoError(t, err)
	assert.Equal(t, map[string]string{"svc-vlagent": "victoriametrics/vlagent:v1.53.0"}, f.updated)
}

// A release that names no logging images does not touch the apps at all -
// including not looking them up.
func TestLoggingUpdateDoesNothingWithoutTargetImages(t *testing.T) {
	f, apps := loggingStack(map[string]string{
		base.HivepaasVictoriaLogsKey: "victoriametrics/victoria-logs:v1.52.0",
	})
	s := &service{dockerManager: f, systemAppService: apps}

	err := s.updateLoggingService(context.Background(), nil, loggingUpdateData(t, &base.ReleaseInfo{}))

	assert.NoError(t, err)
	assert.Empty(t, f.updated)
	assert.Empty(t, apps.recorded)
}

// Swarm's default on a failed update is to stop and leave the service down.
// Every service the updater moves has to be told to undo it instead, and the
// shared step is what guarantees that rather than each caller remembering.
func TestLoggingUpdateArmsSwarmRollback(t *testing.T) {
	f, apps := loggingStack(map[string]string{
		base.HivepaasVictoriaLogsKey: "victoriametrics/victoria-logs:v1.52.0",
	})
	s := &service{dockerManager: f, systemAppService: apps}

	err := s.updateLoggingService(context.Background(), nil, loggingUpdateData(t, &base.ReleaseInfo{
		VictoriaLogsImage: "victoriametrics/victoria-logs:v1.53.0",
	}))

	assert.NoError(t, err)
	spec := f.specs["svc-victoria-logs"]
	if assert.NotNil(t, spec) && assert.NotNil(t, spec.UpdateConfig) {
		assert.Equal(t, swarm.UpdateFailureActionRollback, spec.UpdateConfig.FailureAction)
		assert.InDelta(t, updateMaxFailureRatio, spec.UpdateConfig.MaxFailureRatio, 0.0001)
	}
}
