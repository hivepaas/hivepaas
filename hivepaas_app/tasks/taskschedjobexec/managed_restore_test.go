package taskschedjobexec

import (
	"context"
	"testing"
	"time"

	"github.com/moby/moby/api/types/swarm"
	"github.com/stretchr/testify/assert"

	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/infra/database"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/tasklog"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/schedjobexecservice"
	"github.com/hivepaas/hivepaas/hivepaas_app/tasks/queue"
	"github.com/hivepaas/hivepaas/services/docker"
)

// scalingDocker scales the services it is asked to, and records the replicas it
// was given, in order.
type scalingDocker struct {
	docker.Manager
	replicas []uint64
}

func (d *scalingDocker) ServiceUpdateFunc(_ context.Context, _ string, _ *swarm.Service,
	update func(int, *swarm.Service) (bool, error), _ int, _ time.Duration, _ ...docker.ServiceUpdateOption) error {
	svc := &swarm.Service{Spec: swarm.ServiceSpec{Mode: swarm.ServiceMode{
		Replicated: &swarm.ReplicatedService{Replicas: new(uint64(0))}}}}
	if _, err := update(0, svc); err != nil {
		return err
	}
	d.replicas = append(d.replicas, *svc.Spec.Mode.Replicated.Replicas)
	return nil
}

// verifierExec runs the restore test in the verifier, and records where.
type verifierExec struct {
	schedjobexecservice.Service
	destApp string
	waited  bool
	err     error
}

func (f *verifierExec) SchedJobExec(_ context.Context, _ database.Tx,
	req *schedjobexecservice.SchedJobExecReq) (*schedjobexecservice.SchedJobExecResp, error) {
	f.destApp = req.DestApp.ID
	f.waited = req.TaskFindRetryMax > 0
	exitCode := 0
	return &schedjobexecservice.SchedJobExecResp{ExitCode: &exitCode,
		Outputs: map[string]string{"DURATION_SECONDS": "95"}}, f.err
}

func verifierRun() *jobRun {
	job := &entity.SchedJob{JobType: base.SchedJobTypeManagedRestoreTest, App: entity.ObjectID{ID: "ver-1"},
		Command:   &entity.CommandTemplate{Command: "/opt/hivepaas/bin/walg-restore-test"},
		ManagedBy: entity.ObjectID{ID: "m1"}}
	setting := &entity.Setting{ID: "j1", Type: base.SettingTypeSchedJob}
	setting.MustSetData(job)
	refs := entity.NewRefObjects()
	refs.RefApps["ver-1"] = &entity.App{ID: "ver-1", ServiceID: "svc-ver-1"}
	return &jobRun{
		execData:   &queue.TaskExecData{Task: &entity.Task{ID: "t1"}, LogStore: tasklog.NewNullStore()},
		jobSetting: setting, refObjects: refs,
	}
}

// The verifier is up only for its test: scaled to one, the test run in it once
// a task of it runs, and scaled back to none - also when the test fails. No
// license is consulted (the executor has none here): a lapsed plan still tests.
func TestARestoreTestRunsInTheVerifierForItsTimeOnly(t *testing.T) {
	for name, execErr := range map[string]error{"passing": nil, "failing": hperrors.Wrap(hperrors.ErrInfraActionFailed)} {
		docker := &scalingDocker{}
		exec := &verifierExec{err: execErr}
		e := &Executor{dockerManager: docker, schedJobExecService: exec}

		result, err := e.runJob(context.Background(), database.Tx{}, verifierRun())

		if execErr != nil {
			assert.Error(t, err, name)
		} else {
			assert.NoError(t, err, name)
		}
		assert.Equal(t, []uint64{1, 0}, docker.replicas, name)
		assert.Equal(t, "ver-1", exec.destApp, name)
		assert.True(t, exec.waited, "%s: waits for the verifier's task to start", name)
		if assert.NotNil(t, result, name) {
			assert.Equal(t, "95", result.outputs["DURATION_SECONDS"], name)
		}
	}
}

func TestARestoreTestWithoutItsVerifierFailsSayingSo(t *testing.T) {
	run := verifierRun()
	run.refObjects = entity.NewRefObjects()

	_, err := (&Executor{dockerManager: &scalingDocker{}, schedJobExecService: &verifierExec{}}).
		runJob(context.Background(), database.Tx{}, run)

	assert.ErrorIs(t, err, hperrors.ErrNotFound)
}
