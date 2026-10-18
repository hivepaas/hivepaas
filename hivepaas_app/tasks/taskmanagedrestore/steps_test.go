package taskmanagedrestore

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
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/bunex"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/tasklog"
	"github.com/hivepaas/hivepaas/hivepaas_app/repository"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/envvarservice"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/managedservice/postgres"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/schedjobexecservice"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/schedjobservice"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/settingmountservice"
	"github.com/hivepaas/hivepaas/hivepaas_app/tasks/queue"
	"github.com/hivepaas/hivepaas/services/docker"
)

var start = time.Date(2026, 10, 18, 15, 0, 0, 0, time.UTC)

// fakeExec answers each recovery check with the next exit code; nil is a check
// that found no task to run in.
type fakeExec struct {
	schedjobexecservice.Service
	exits    []*int
	commands []string
}

func (f *fakeExec) RunCommand(_ context.Context, _ database.Tx, req *schedjobexecservice.RunCommandReq) (
	*schedjobexecservice.RunCommandResp, error) {
	f.commands = append(f.commands, req.Command.Command)
	exit := f.exits[0]
	if len(f.exits) > 1 {
		f.exits = f.exits[1:]
	}
	if exit == nil {
		return &schedjobexecservice.RunCommandResp{}, hperrors.NewNotFound("App task")
	}
	return &schedjobexecservice.RunCommandResp{ExitCode: exit}, nil
}

type fakeTasks struct {
	repository.TaskRepo
	inserted []*entity.Task
	byID     map[string]*entity.Task
}

func (f *fakeTasks) Insert(_ context.Context, _ database.IDB, task *entity.Task, _ ...bunex.InsertQueryOption) error {
	f.inserted = append(f.inserted, task)
	return nil
}

func (f *fakeTasks) GetByID(_ context.Context, _ database.IDB, _ *entity.ObjectScope, _ base.TaskType, id string,
	_ ...bunex.SelectQueryOption) (*entity.Task, error) {
	if task := f.byID[id]; task != nil {
		return task, nil
	}
	return nil, hperrors.NewNotFound("Task")
}

type fakeSettings struct {
	repository.SettingRepo
	byID    map[string]*entity.Setting
	listed  []*entity.Setting
	updated []*entity.Setting
}

func (f *fakeSettings) GetByID(_ context.Context, _ database.IDB, _ *entity.ObjectScope, _ base.SettingType,
	id string, _ bool, _ ...bunex.SelectQueryOption) (*entity.Setting, error) {
	return f.byID[id], nil
}

func (f *fakeSettings) List(_ context.Context, _ database.IDB, _ *entity.ObjectScope, _ *basedto.Paging,
	_ ...bunex.SelectQueryOption) ([]*entity.Setting, *basedto.PagingMeta, error) {
	return f.listed, nil, nil
}

func (f *fakeSettings) Update(_ context.Context, _ database.IDB, setting *entity.Setting,
	_ ...bunex.UpdateQueryOption) error {
	f.updated = append(f.updated, setting)
	return nil
}

type fakeSchedJobs struct {
	schedjobservice.Service
}

func (f *fakeSchedJobs) CreateSchedJobTask(job *entity.Setting, runAt, _ time.Time) (*entity.Task, error) {
	return &entity.Task{ID: "run-1", Type: base.TaskTypeSchedJobExec, TargetID: job.ID, RunAt: runAt}, nil
}

type fakeQueue struct {
	queue.TaskQueue
	scheduled []*entity.Task
	jobs      []*entity.Setting
}

func (f *fakeQueue) ScheduleTasksForSchedJob(_ context.Context, _ database.Tx, job *entity.Setting, _ bool) error {
	f.jobs = append(f.jobs, job)
	return nil
}

func (f *fakeQueue) ScheduleTask(_ context.Context, tasks ...*entity.Task) error {
	f.scheduled = append(f.scheduled, tasks...)
	return nil
}

type fakeEnv struct {
	envvarservice.Service
	applied []*envvarservice.AppEnvVarData
}

func (f *fakeEnv) BuildEnvVarsForAllAppsInScope(_ context.Context, _ database.IDB, _ *entity.ObjectScope, _ bool,
	onlyApps []string, _, _ bool) ([]*envvarservice.AppEnvVarData, error) {
	return []*envvarservice.AppEnvVarData{{App: &entity.App{ID: onlyApps[0]}}}, nil
}

func (f *fakeEnv) ApplyEnvVarsForApps(_ context.Context, _ database.IDB, data []*envvarservice.AppEnvVarData,
	_, _ bool) map[int]error {
	f.applied = append(f.applied, data...)
	return nil
}

type fakeMounts struct {
	settingmountservice.Service
	refreshed []*entity.Setting
}

func (f *fakeMounts) RecordRefresh(_ context.Context, _ database.IDB, settings ...*entity.Setting) (
	*entity.Task, error) {
	f.refreshed = append(f.refreshed, settings...)
	return &entity.Task{ID: "refresh-1"}, nil
}

type fakeDocker struct {
	docker.Manager
	service *swarm.Service
}

func (f *fakeDocker) ServiceUpdateFunc(_ context.Context, _ string, _ *swarm.Service,
	update func(int, *swarm.Service) (bool, error), _ int, _ time.Duration, _ ...docker.ServiceUpdateOption) error {
	_, err := update(0, f.service)
	return err
}

// testExecutor's clock moves only when it sleeps.
func testExecutor() (*Executor, *time.Time) {
	now := start
	return &Executor{
		now: func() time.Time { return now },
		sleep: func(_ context.Context, d time.Duration) error {
			now = now.Add(d)
			return nil
		},
	}, &now
}

func exit(code int) *int { return &code }

func testRun(step string) *restoreRun {
	managed := &entity.AppManagedSettings{Engine: "postgres", Postgres: &entity.ManagedPostgres{
		Jobs:    entity.ManagedPostgresJobs{Nightly: entity.ObjectID{ID: "job-nightly"}},
		Restore: &entity.ManagedRestore{SourceApp: entity.ObjectID{ID: "app-src"}},
	}}
	setting := &entity.Setting{ID: "managed-1", Type: base.SettingTypeAppManaged, UpdateVer: 1}
	setting.MustSetData(managed)
	return &restoreRun{
		data:    &queue.TaskExecData{Task: &entity.Task{ID: "restore-1"}, LogStore: tasklog.NewNullStore()},
		app:     &entity.App{ID: "app-new", ServiceID: "svc-new"},
		setting: setting, managed: managed, engine: postgres.New(),
		state: &entity.TaskManagedRestoreState{Step: step},
	}
}

// A recovery takes as long as the WAL it replays: a run waits its slice on it,
// and continues the task, which asks again.
func TestARecoveryIsWaitedForAcrossRuns(t *testing.T) {
	e, _ := testExecutor()
	exec := &fakeExec{exits: []*int{nil, exit(75)}}
	e.schedJobExecService = exec
	run := testRun(entity.ManagedRestoreStepRecover)

	assert.NoError(t, e.runStep(context.Background(), database.Tx{}, run))

	assert.True(t, run.data.Continued())
	assert.Equal(t, entity.ManagedRestoreStepRecover, run.state.Step, "still recovering")
	assert.Equal(t, start, run.state.RecoverySince)
	assert.Equal(t, "/opt/hivepaas/bin/walg-entrypoint status", exec.commands[0])
	assert.Greater(t, len(exec.commands), 2, "asked again within the run")
	assert.False(t, run.data.TaskNonRetryable)
}

func TestARecoveredDatabaseGoesOnToItsFirstBackup(t *testing.T) {
	e, _ := testExecutor()
	e.schedJobExecService = &fakeExec{exits: []*int{exit(75), exit(0)}}
	run := testRun(entity.ManagedRestoreStepRecover)

	assert.NoError(t, e.runStep(context.Background(), database.Tx{}, run))

	assert.Equal(t, entity.ManagedRestoreStepBackup, run.state.Step)
	assert.True(t, run.data.Continued())
}

// A recovery that ended before its target fails the restore with Postgres's own
// message, which the check printed into the task's log; it is not tried again.
func TestAFailedRecoveryEndsTheRestore(t *testing.T) {
	e, _ := testExecutor()
	e.schedJobExecService = &fakeExec{exits: []*int{exit(1)}}
	run := testRun(entity.ManagedRestoreStepRecover)

	err := e.runStep(context.Background(), database.Tx{}, run)

	assert.ErrorIs(t, err, hperrors.ErrAppManagedRestoreFailed)
	assert.True(t, run.data.TaskNonRetryable)
	assert.False(t, run.data.Continued())
}

func TestARecoveryPastItsDeadlineEndsTheRestore(t *testing.T) {
	e, _ := testExecutor()
	e.schedJobExecService = &fakeExec{exits: []*int{exit(75)}}
	run := testRun(entity.ManagedRestoreStepRecover)
	run.state.RecoverySince = start.Add(-recoveryDeadline - time.Minute)

	assert.ErrorIs(t, e.runStep(context.Background(), database.Tx{}, run), hperrors.ErrAppManagedRestoreFailed)
	assert.True(t, run.data.TaskNonRetryable)
}

// The engine's jobs start once the database has recovered, and the first base
// backup is a run of its nightly job, which Status reads as any other; the run
// is scheduled once this run's transaction commits.
func TestTheFirstBackupIsARunOfTheNightlyJob(t *testing.T) {
	e, _ := testExecutor()
	job := func(id, name string) *entity.Setting {
		setting := &entity.Setting{ID: id, Type: base.SettingTypeSchedJob, Name: name,
			Status: base.SettingStatusDisabled}
		setting.MustSetData(&entity.SchedJob{JobType: base.SchedJobTypeContainerCommand,
			ManagedBy: entity.ObjectID{ID: "managed-1"}})
		return setting
	}
	nightly, health := job("job-nightly", "Nightly base backup"), job("job-health", "Archive health")
	mine := &entity.Setting{ID: "job-mine", Type: base.SettingTypeSchedJob, Status: base.SettingStatusDisabled}
	mine.MustSetData(&entity.SchedJob{JobType: base.SchedJobTypeContainerCommand})
	settings, tasks, q := &fakeSettings{listed: []*entity.Setting{nightly, health, mine}}, &fakeTasks{}, &fakeQueue{}
	e.settingRepo, e.taskRepo, e.taskQueue = settings, tasks, q
	e.schedJobService = &fakeSchedJobs{}
	run := testRun(entity.ManagedRestoreStepBackup)

	assert.NoError(t, e.runStep(context.Background(), database.Tx{}, run))

	assert.Equal(t, base.SettingStatusActive, nightly.Status)
	assert.Equal(t, base.SettingStatusActive, health.Status)
	assert.Equal(t, base.SettingStatusDisabled, mine.Status, "a person's job is theirs")
	assert.ElementsMatch(t, []*entity.Setting{nightly, health}, q.jobs, "scheduled as they start")
	if assert.Len(t, tasks.inserted, 1) {
		assert.Equal(t, "job-nightly", tasks.inserted[0].TargetID)
	}
	assert.Equal(t, "run-1", run.state.BackupTask)
	assert.Equal(t, entity.ManagedRestoreStepWaitBackup, run.state.Step)
	assert.Empty(t, q.scheduled, "not before the commit")
	run.data.OnPostTxFunc()
	assert.Equal(t, tasks.inserted, q.scheduled)
}

func TestTheBackupsEndDecidesTheNextStep(t *testing.T) {
	for name, tc := range map[string]struct {
		status    base.TaskStatus
		retryLeft bool
		step      string
		err       error
	}{
		"done":                  {base.TaskStatusDone, false, entity.ManagedRestoreStepFinish, nil},
		"running":               {base.TaskStatusInProgress, false, entity.ManagedRestoreStepWaitBackup, nil},
		"failed, to be retried": {base.TaskStatusFailed, true, entity.ManagedRestoreStepWaitBackup, nil},
		"failed for good": {base.TaskStatusFailed, false, entity.ManagedRestoreStepWaitBackup,
			hperrors.ErrAppManagedRestoreFailed},
		"canceled": {base.TaskStatusCanceled, false, entity.ManagedRestoreStepWaitBackup,
			hperrors.ErrAppManagedRestoreFailed},
	} {
		e, _ := testExecutor()
		backup := &entity.Task{ID: "run-1", Status: tc.status}
		if tc.retryLeft {
			backup.Config.MaxRetry = 2
		}
		e.taskRepo = &fakeTasks{byID: map[string]*entity.Task{"run-1": backup}}
		run := testRun(entity.ManagedRestoreStepWaitBackup)
		run.state.BackupTask = "run-1"

		err := e.runStep(context.Background(), database.Tx{}, run)

		if tc.err != nil {
			assert.ErrorIs(t, err, tc.err, name)
			continue
		}
		assert.NoError(t, err, name)
		assert.Equal(t, tc.step, run.state.Step, name)
		assert.True(t, run.data.Continued(), name)
	}
}

// Once ready, the app keeps nothing of the source: the variables go, in its
// settings and on its service; the keys go, and are unmounted by a refresh; the
// source's mount comes off; the engine's setting says it is ready.
func TestFinishingTakesOffWhatTheAppWasRestoredWith(t *testing.T) {
	e, now := testExecutor()
	env := &entity.Setting{ID: "env-1", Type: base.SettingTypeEnvVar, ObjectID: "app-new"}
	env.MustSetData(&entity.EnvVars{Data: []*entity.EnvVar{{Key: "WALG_S3_PREFIX", Value: "s3://b/new"},
		{Key: "HIVEPAAS_RESTORE_WALG_FILE_PREFIX", Value: "/var/lib/hivepaas-walg-source"}}})
	key := &entity.Setting{ID: "sec-2", Type: base.SettingTypeSecret, Name: "RESTORE_WALG_LIBSODIUM_KEY",
		ObjectID: "app-new"}
	key.MustSetData(&entity.Secret{Key: "RESTORE_WALG_LIBSODIUM_KEY"})
	keyMount := &entity.Setting{ID: "mnt-2", Type: base.SettingTypeAppSettingMount, ObjectID: "app-new"}
	keyMount.MustSetData(&entity.AppSettingMount{Source: entity.ObjectID{ID: "sec-2"}})
	settings := &fakeSettings{listed: []*entity.Setting{env, key, keyMount}}
	envs, mounts, q := &fakeEnv{}, &fakeMounts{}, &fakeQueue{}
	service := &swarm.Service{Spec: swarm.ServiceSpec{TaskTemplate: swarm.TaskSpec{ContainerSpec: &swarm.ContainerSpec{
		Mounts: []mount.Mount{{Target: "/var/lib/postgresql"}, {Target: "/var/lib/hivepaas-walg-source"}}}}}}
	e.settingRepo, e.envVarService, e.settingMountService, e.taskQueue = settings, envs, mounts, q
	e.dockerManager = &fakeDocker{service: service}
	run := testRun(entity.ManagedRestoreStepFinish)
	run.managed.Locks.Mounts = []string{"/var/lib/postgresql", "/var/lib/hivepaas-walg-source"}

	assert.NoError(t, e.runStep(context.Background(), database.Tx{}, run))

	assert.Equal(t, entity.ManagedRestoreStepDone, run.state.Step)
	assert.False(t, run.data.Continued(), "nothing left to do")
	updated := map[string]*entity.Setting{}
	for _, s := range settings.updated {
		updated[s.ID] = s
	}
	assert.Len(t, updated["env-1"].MustAsEnvVars().Data, 1)
	assert.Equal(t, *now, updated["sec-2"].DeletedAt)
	assert.Equal(t, *now, updated["mnt-2"].DeletedAt)
	ready := updated["managed-1"].MustAsAppManagedSettings()
	assert.Equal(t, *now, ready.Postgres.Restore.ReadyAt)
	assert.Equal(t, []string{"/var/lib/postgresql"}, ready.Locks.Mounts)
	if assert.Len(t, envs.applied, 1) {
		assert.Equal(t, "app-new", envs.applied[0].App.ID)
	}
	assert.Equal(t, []mount.Mount{{Target: "/var/lib/postgresql"}}, service.Spec.TaskTemplate.ContainerSpec.Mounts)
	assert.ElementsMatch(t, []string{"sec-2", "mnt-2"}, []string{mounts.refreshed[0].ID, mounts.refreshed[1].ID})
	run.data.OnPostTxFunc()
	if assert.Len(t, q.scheduled, 1) {
		assert.Equal(t, "refresh-1", q.scheduled[0].ID)
	}
}

// A run after the last step does nothing: a crash after it, or a stray run.
func TestADoneRestoreDoesNothing(t *testing.T) {
	e, _ := testExecutor()
	run := testRun(entity.ManagedRestoreStepDone)

	assert.NoError(t, e.runStep(context.Background(), database.Tx{}, run))
	assert.False(t, run.data.Continued())
}
