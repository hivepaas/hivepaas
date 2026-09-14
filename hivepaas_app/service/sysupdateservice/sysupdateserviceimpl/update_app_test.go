package sysupdateserviceimpl

import (
	"context"
	"testing"

	"github.com/moby/moby/api/types/swarm"
	"github.com/stretchr/testify/assert"

	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/tasklog"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/hpappservice"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/sysupdateservice"
	"github.com/hivepaas/hivepaas/hivepaas_app/tasks/queue"
)

// fakeHpApp answers for the app and worker services. Embedding the interface
// means only what this path calls needs a body.
type fakeHpApp struct {
	hpappservice.Service
	app       *swarm.Service
	db        *swarm.Service
	worker    *swarm.Service
	workerErr error
}

func (f *fakeHpApp) GetHpDbSwarmService(_ context.Context) (*swarm.Service, error) {
	return f.db, nil
}

func (f *fakeHpApp) GetHpAppSwarmService(_ context.Context) (*swarm.Service, error) {
	return f.app, nil
}

func (f *fakeHpApp) GetHpWorkerSwarmService(_ context.Context) (*swarm.Service, error) {
	if f.workerErr != nil {
		return nil, f.workerErr
	}
	if f.worker == nil {
		return nil, hperrors.Wrap(hperrors.ErrNotFound)
	}
	return f.worker, nil
}

func swarmServiceOn(id, image string, replicas uint64) *swarm.Service {
	return &swarm.Service{
		ID: id,
		Spec: swarm.ServiceSpec{
			Annotations:  swarm.Annotations{Name: id},
			TaskTemplate: swarm.TaskSpec{ContainerSpec: &swarm.ContainerSpec{Image: image}},
			Mode:         swarm.ServiceMode{Replicated: &swarm.ReplicatedService{Replicas: &replicas}},
		},
	}
}

// Swarm's default on a failed update is to stop and leave the service down. The
// app and the worker went without this for a while, which made the rollback the
// rest of the update path relies on unreachable for exactly the two services
// whose failure is worst.
func TestAppAndWorkerUpdatesArmSwarmRollback(t *testing.T) {
	f := &fakeDocker{}
	hp := &fakeHpApp{
		app:    swarmServiceOn("hivepaas_app", "hivepaas/hivepaas-dev:0.1.0", 1),
		worker: swarmServiceOn("hivepaas_worker", "hivepaas/hivepaas-dev:0.1.0", 2),
	}
	s := &service{dockerManager: f, hpAppService: hp}
	data := loggingUpdateData(t, &base.ReleaseInfo{AppImage: "hivepaas/hivepaas-dev:0.2.0"})
	data.CurrentAppReplicas = new(uint64(1))
	data.CurrentWorkerReplicas = new(uint64(2))

	assert.NoError(t, s.updateMainAppService(context.Background(), data))
	assert.NoError(t, s.updateWorkerService(context.Background(), data))

	for _, name := range []string{"hivepaas_app", "hivepaas_worker"} {
		spec := f.specs[name]
		if assert.NotNil(t, spec, name) && assert.NotNil(t, spec.UpdateConfig, name) {
			assert.Equal(t, swarm.UpdateFailureActionRollback, spec.UpdateConfig.FailureAction, name)
			assert.InDelta(t, updateMaxFailureRatio, spec.UpdateConfig.MaxFailureRatio, 0.0001, name)
		}
		assert.Equal(t, "hivepaas/hivepaas-dev:0.2.0", f.updated[name], name)
	}
}

// The replicas the update puts back are the ones stopServices took away, not
// whatever the spec happened to hold.
func TestAppUpdateRestoresTheReplicasItStoppedWith(t *testing.T) {
	f := &fakeDocker{}
	hp := &fakeHpApp{app: swarmServiceOn("hivepaas_app", "hivepaas/hivepaas-dev:0.1.0", 0)}
	s := &service{dockerManager: f, hpAppService: hp}
	data := loggingUpdateData(t, &base.ReleaseInfo{AppImage: "hivepaas/hivepaas-dev:0.2.0"})
	data.CurrentAppReplicas = new(uint64(3))

	assert.NoError(t, s.updateMainAppService(context.Background(), data))

	spec := f.specs["hivepaas_app"]
	if assert.NotNil(t, spec) {
		assert.Equal(t, uint64(3), *spec.Mode.Replicated.Replicas)
	}
}

// Running the same release twice must not restart the app.
func TestAppUpdateLeavesTheSameImageAlone(t *testing.T) {
	f := &fakeDocker{}
	hp := &fakeHpApp{app: swarmServiceOn("hivepaas_app", "hivepaas/hivepaas-dev:0.2.0", 1)}
	s := &service{dockerManager: f, hpAppService: hp}
	data := loggingUpdateData(t, &base.ReleaseInfo{AppImage: "hivepaas/hivepaas-dev:0.2.0"})

	assert.NoError(t, s.updateMainAppService(context.Background(), data))

	assert.Empty(t, f.updated)
}

// An installation running the worker inside the main app has no worker service.
// That is a configuration, not a failure.
func TestWorkerUpdateSkipsAnInstallationWithoutOne(t *testing.T) {
	f := &fakeDocker{}
	s := &service{dockerManager: f, hpAppService: &fakeHpApp{
		app: swarmServiceOn("hivepaas_app", "hivepaas/hivepaas-dev:0.1.0", 1),
	}}
	data := loggingUpdateData(t, &base.ReleaseInfo{AppImage: "hivepaas/hivepaas-dev:0.2.0"})

	assert.NoError(t, s.updateWorkerService(context.Background(), data))

	assert.Empty(t, f.updated)
}

// The update's own context carries a one hour deadline, and the step that puts
// the app back runs after the update has ended - including when it ended because
// that deadline passed. Restoring through the expired context would do nothing
// and leave the installation stopped with no dashboard to fix it from, so the
// restore gets a context of its own.
func TestRestoreRunsEvenWhenTheUpdateContextIsDead(t *testing.T) {
	f := &fakeDocker{}
	s := &service{dockerManager: f, hpAppService: &fakeHpApp{
		app: swarmServiceOn("hivepaas_app", "hivepaas/hivepaas-dev:0.1.0", 2),
		// Fails the update immediately after the app has been stopped, which is
		// the window this is about.
		workerErr: hperrors.Wrap(hperrors.ErrInfra),
	}}

	task := &entity.Task{ID: "t1", Type: base.TaskTypeSystemUpdate}
	task.Config.MaxRetry = 1 // keeps the result notification out of this test
	task.MustSetArgs(&entity.TaskSystemUpdateArgs{TargetVersion: &base.ReleaseInfo{}})
	req := &sysupdateservice.SysUpdateReq{
		TaskExecData: &queue.TaskExecData{Task: task, LogStore: tasklog.NewLocalStore("t1")},
	}

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	_, err := s.SysUpdate(ctx, nil, req)

	assert.Error(t, err, "the update itself still fails")
	if assert.Len(t, f.replicaUpdates, 2, "stopped once, restored once") {
		assert.Equal(t, uint64(0), f.replicaUpdates[0].replicas, "stopped")
		assert.Error(t, f.replicaUpdates[0].ctxErr, "the update ran on the dead context")

		assert.Equal(t, uint64(2), f.replicaUpdates[1].replicas, "restored to what it was")
		assert.NoError(t, f.replicaUpdates[1].ctxErr, "the restore did not")
	}
}

// The replica count is compared inside the callback now, against the service as
// it is at that moment rather than against whatever the caller fetched earlier.
// Asking for the count it already has sends nothing.
func TestScalingToTheCountItAlreadyHasSendsNothing(t *testing.T) {
	f := &fakeDocker{}
	s := &service{dockerManager: f}

	err := s.scaleServiceReplicas(context.Background(), swarmServiceOn("svc", "img:1", 2), 2)

	assert.NoError(t, err)
	assert.Empty(t, f.replicaUpdates)
}

func TestScalingToADifferentCountSendsIt(t *testing.T) {
	f := &fakeDocker{}
	s := &service{dockerManager: f}

	err := s.scaleServiceReplicas(context.Background(), swarmServiceOn("svc", "img:1", 2), 0)

	assert.NoError(t, err)
	if assert.Len(t, f.replicaUpdates, 1) {
		assert.Equal(t, uint64(0), f.replicaUpdates[0].replicas)
	}
}

// A global service has no replica count to set, and reading one would panic.
func TestScalingIgnoresANonReplicatedService(t *testing.T) {
	f := &fakeDocker{}
	s := &service{dockerManager: f}

	err := s.scaleServiceReplicas(context.Background(), &swarm.Service{ID: "svc"}, 1)

	assert.NoError(t, err)
	assert.Empty(t, f.replicaUpdates)
}
