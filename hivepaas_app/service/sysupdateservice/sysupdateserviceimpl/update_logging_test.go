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
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/tasklog"
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

func TestLoggingUpdateMovesBothServicesToANewerImage(t *testing.T) {
	f := &fakeDocker{deployed: map[string]string{
		base.HivepaasVictoriaLogsServiceName: "victoriametrics/victoria-logs:v1.52.0",
		base.HivepaasVlagentServiceName:      "victoriametrics/vlagent:v1.52.0",
	}}
	s := &service{dockerManager: f}

	err := s.updateLoggingService(context.Background(), loggingUpdateData(t, &base.ReleaseInfo{
		VictoriaLogsImage: "victoriametrics/victoria-logs:v1.53.0",
		VlagentImage:      "victoriametrics/vlagent:v1.53.0",
	}))

	assert.NoError(t, err)
	assert.Equal(t, map[string]string{
		base.HivepaasVictoriaLogsServiceName: "victoriametrics/victoria-logs:v1.53.0",
		base.HivepaasVlagentServiceName:      "victoriametrics/vlagent:v1.53.0",
	}, f.updated)
}

// Re-running the same release must not restart the logging stack. This is the
// whole point of the version check.
func TestLoggingUpdateLeavesTheSameVersionAlone(t *testing.T) {
	f := &fakeDocker{deployed: map[string]string{
		base.HivepaasVictoriaLogsServiceName: "victoriametrics/victoria-logs:v1.52.0",
		base.HivepaasVlagentServiceName:      "victoriametrics/vlagent:v1.52.0",
	}}
	s := &service{dockerManager: f}

	err := s.updateLoggingService(context.Background(), loggingUpdateData(t, &base.ReleaseInfo{
		VictoriaLogsImage: "victoriametrics/victoria-logs:v1.52.0",
		VlagentImage:      "victoriametrics/vlagent:v1.52.0",
	}))

	assert.NoError(t, err)
	assert.Empty(t, f.updated)
}

// A digest is what a running service actually reports, and it moves whenever a
// tag is re-pushed. Same tag is still the same version.
func TestLoggingUpdateIgnoresADigestUnderTheSameTag(t *testing.T) {
	f := &fakeDocker{deployed: map[string]string{
		base.HivepaasVictoriaLogsServiceName: "victoriametrics/victoria-logs:v1.52.0@sha256:aaaa",
	}}
	s := &service{dockerManager: f}

	err := s.updateLoggingService(context.Background(), loggingUpdateData(t, &base.ReleaseInfo{
		VictoriaLogsImage: "victoriametrics/victoria-logs:v1.52.0",
	}))

	assert.NoError(t, err)
	assert.Empty(t, f.updated)
}

func TestLoggingUpdateRefusesToGoBackwards(t *testing.T) {
	f := &fakeDocker{deployed: map[string]string{
		base.HivepaasVictoriaLogsServiceName: "victoriametrics/victoria-logs:v1.52.0",
	}}
	s := &service{dockerManager: f}

	err := s.updateLoggingService(context.Background(), loggingUpdateData(t, &base.ReleaseInfo{
		VictoriaLogsImage: "victoriametrics/victoria-logs:v1.51.0",
	}))

	assert.NoError(t, err)
	assert.Empty(t, f.updated)
}

// Logging is optional. Neither service existing is the ordinary state of an
// install that never switched it on, and it must not fail the whole update.
func TestLoggingUpdateSkipsWhatIsNotDeployed(t *testing.T) {
	f := &fakeDocker{deployed: map[string]string{}}
	s := &service{dockerManager: f}

	err := s.updateLoggingService(context.Background(), loggingUpdateData(t, &base.ReleaseInfo{
		VictoriaLogsImage: "victoriametrics/victoria-logs:v1.53.0",
		VlagentImage:      "victoriametrics/vlagent:v1.53.0",
	}))

	assert.NoError(t, err)
	assert.Empty(t, f.updated)
}

// The collector is deployed but the backend is somebody else's: only the one
// that is there moves.
func TestLoggingUpdateMovesOnlyTheServiceThatExists(t *testing.T) {
	f := &fakeDocker{deployed: map[string]string{
		base.HivepaasVlagentServiceName: "victoriametrics/vlagent:v1.52.0",
	}}
	s := &service{dockerManager: f}

	err := s.updateLoggingService(context.Background(), loggingUpdateData(t, &base.ReleaseInfo{
		VictoriaLogsImage: "victoriametrics/victoria-logs:v1.53.0",
		VlagentImage:      "victoriametrics/vlagent:v1.53.0",
	}))

	assert.NoError(t, err)
	assert.Equal(t, map[string]string{
		base.HivepaasVlagentServiceName: "victoriametrics/vlagent:v1.53.0",
	}, f.updated)
}

// A release that names no logging images does not touch the services at all -
// including not inspecting them.
func TestLoggingUpdateDoesNothingWithoutTargetImages(t *testing.T) {
	f := &fakeDocker{deployed: map[string]string{
		base.HivepaasVictoriaLogsServiceName: "victoriametrics/victoria-logs:v1.52.0",
	}}
	s := &service{dockerManager: f}

	err := s.updateLoggingService(context.Background(), loggingUpdateData(t, &base.ReleaseInfo{}))

	assert.NoError(t, err)
	assert.Empty(t, f.updated)
}

// Swarm's default on a failed update is to stop and leave the service down.
// Every service the updater moves has to be told to undo it instead, and the
// shared step is what guarantees that rather than each caller remembering.
func TestLoggingUpdateArmsSwarmRollback(t *testing.T) {
	f := &fakeDocker{deployed: map[string]string{
		base.HivepaasVictoriaLogsServiceName: "victoriametrics/victoria-logs:v1.52.0",
	}}
	s := &service{dockerManager: f}

	err := s.updateLoggingService(context.Background(), loggingUpdateData(t, &base.ReleaseInfo{
		VictoriaLogsImage: "victoriametrics/victoria-logs:v1.53.0",
	}))

	assert.NoError(t, err)
	spec := f.specs[base.HivepaasVictoriaLogsServiceName]
	if assert.NotNil(t, spec) && assert.NotNil(t, spec.UpdateConfig) {
		assert.Equal(t, swarm.UpdateFailureActionRollback, spec.UpdateConfig.FailureAction)
		assert.InDelta(t, updateMaxFailureRatio, spec.UpdateConfig.MaxFailureRatio, 0.0001)
	}
}
