package taskschedjobexec

import (
	"context"
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/infra/database"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/tasklog"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/apptemplateservice/templatemodel"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/managedservice"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/managedservice/managedmodel"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/managedservice/postgres"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/schedjobexecservice"
	"github.com/hivepaas/hivepaas/hivepaas_app/tasks/queue"
)

type fakeManaged struct {
	managedservice.Service
	// pruned are the outputs each prune was given, and prunedApps its apps.
	pruned     []map[string]string
	prunedApps []string
}

func (f *fakeManaged) PruneDeployPoints(_ context.Context, _ database.IDB, _, appID string,
	outputs map[string]string) ([]string, error) {
	f.pruned = append(f.pruned, outputs)
	f.prunedApps = append(f.prunedApps, appID)
	return nil, nil
}

func (f *fakeManaged) Engine(name templatemodel.ManagedEngine) (managedmodel.Engine, bool) {
	return postgres.New(), name == templatemodel.ManagedEnginePostgres
}

// nightlyExec runs each step, answering with the outputs given for it.
type nightlyExec struct {
	schedjobexecservice.Service
	commands []string
	// apps are the apps each step ran in.
	apps    []string
	outputs map[string]map[string]string
	failing string
}

func (f *nightlyExec) SchedJobExec(_ context.Context, _ database.Tx,
	req *schedjobexecservice.SchedJobExecReq) (*schedjobexecservice.SchedJobExecResp, error) {
	f.commands = append(f.commands, req.Command.Command)
	f.apps = append(f.apps, req.DestApp.ID)
	step := strings.Fields(req.Command.Command)[1]
	exitCode := 0
	resp := &schedjobexecservice.SchedJobExecResp{ExitCode: &exitCode, Outputs: f.outputs[step]}
	if step == f.failing {
		exitCode = 1
		return resp, hperrors.Wrap(hperrors.ErrInfraActionFailed)
	}
	return resp, nil
}

// kept backups, a night apart, the last one tonight.
func keptBackups(nights int) string {
	var out []string
	for i := range nights {
		at := time.Now().UTC().AddDate(0, 0, i-nights+1).Format(time.RFC3339)
		out = append(out, fmt.Sprintf("base_%024X@%s@100", i+1, at))
	}
	return strings.Join(out, " ")
}

func nightlyRun(points ...*entity.ManagedRestorePoint) *jobRun {
	job := &entity.SchedJob{JobType: base.SchedJobTypeManagedNightly, App: entity.ObjectID{ID: "db-1"},
		Command:   &entity.CommandTemplate{Command: "/opt/hivepaas/bin/walg-nightly"},
		ManagedBy: entity.ObjectID{ID: "m1"}}
	setting := &entity.Setting{ID: "j1", Type: base.SettingTypeSchedJob, Scope: base.ObjectScopeApp, ObjectID: "db-1"}
	setting.MustSetData(job)
	managed := &entity.Setting{ID: "m1", Type: base.SettingTypeAppManaged}
	managed.MustSetData(&entity.AppManagedSettings{Engine: "postgres",
		Postgres: &entity.ManagedPostgres{RetainFull: 7, RestorePoints: points}})
	refs := entity.NewRefObjects()
	refs.RefApps["db-1"] = &entity.App{ID: "db-1"}
	refs.RefSettings["m1"] = managed
	return &jobRun{execData: &queue.TaskExecData{Task: &entity.Task{ID: "t1"}, LogStore: tasklog.NewNullStore()},
		jobSetting: setting, refObjects: refs}
}

// The retention between the steps is the engine's: here, a point kept from
// eight nights ago holds back the backups before the seventh newest.
func TestANightlyRunBacksUpThenKeepsWhatTheEngineSays(t *testing.T) {
	exec := &nightlyExec{outputs: map[string]map[string]string{
		"backup": {"BACKUP_NAME": "base_00000000000000000000000A", "BACKUPS": keptBackups(10)},
		"retain": {"OLDEST_BACKUP_AT": "2026-10-10T02:00:00Z", "BACKUPS": "after-retention"},
	}}
	kept := &entity.ManagedRestorePoint{Name: "p", LSN: "0/1", At: time.Now().UTC().AddDate(0, 0, -8),
		KeepUntil: time.Now().UTC().AddDate(0, 0, 30)}
	managed := &fakeManaged{}
	e := &Executor{schedJobExecService: exec, managedService: managed}

	result, err := e.runJob(context.Background(), database.Tx{}, nightlyRun(kept))

	assert.NoError(t, err)
	if assert.Len(t, managed.pruned, 1, "the deployments' points are pruned after retention") {
		assert.Equal(t, "after-retention", managed.pruned[0]["BACKUPS"], "from what retention left")
	}
	assert.Equal(t, []string{"/opt/hivepaas/bin/walg-nightly backup",
		"/opt/hivepaas/bin/walg-nightly retain base_000000000000000000000002"}, exec.commands)
	assert.Equal(t, "base_00000000000000000000000A", result.outputs["BACKUP_NAME"], "the backup's outputs")
	assert.Equal(t, "after-retention", result.outputs["BACKUPS"], "the retention's over the backup's")
	assert.Equal(t, "2026-10-10T02:00:00Z", result.outputs["OLDEST_BACKUP_AT"])
}

func TestAFailedBackupIsNotFollowedByRetention(t *testing.T) {
	exec := &nightlyExec{failing: "backup"}
	managed := &fakeManaged{}
	e := &Executor{schedJobExecService: exec, managedService: managed}

	result, err := e.runJob(context.Background(), database.Tx{}, nightlyRun())

	assert.Error(t, err)
	assert.Empty(t, managed.pruned, "nor by pruning")
	assert.Equal(t, []string{"/opt/hivepaas/bin/walg-nightly backup"}, exec.commands)
	if assert.NotNil(t, result) && assert.NotNil(t, result.exitCode) {
		assert.Equal(t, 1, *result.exitCode)
	}
}

func TestANightlyRunWithoutItsEngineFailsSayingSo(t *testing.T) {
	run := nightlyRun()
	run.refObjects.RefSettings = map[string]*entity.Setting{}

	_, err := (&Executor{schedJobExecService: &nightlyExec{}, managedService: &fakeManaged{}}).
		runJob(context.Background(), database.Tx{}, run)

	assert.ErrorIs(t, err, hperrors.ErrNotFound)
}

// A nightly run whose backups are taken on a replica runs its backup there, and
// its retention and pruning on the database the job is the engine's of.
func TestANightlyRunBacksUpOnTheReplicaItNames(t *testing.T) {
	exec := &nightlyExec{outputs: map[string]map[string]string{
		"backup": {"BACKUP_NAME": "base_000000000000000000000002", "BACKUPS": keptBackups(2)},
	}}
	managed := &fakeManaged{}
	e := &Executor{schedJobExecService: exec, managedService: managed}
	run := nightlyRun()
	job := run.jobSetting.MustAsSchedJob()
	job.App = entity.ObjectID{ID: "db-1-r1"}
	run.jobSetting.MustSetData(job)
	run.refObjects.RefApps["db-1-r1"] = &entity.App{ID: "db-1-r1", LogicalParentID: "db-1"}

	_, err := e.runJob(context.Background(), database.Tx{}, run)

	assert.NoError(t, err)
	assert.Equal(t, []string{"db-1-r1", "db-1"}, exec.apps, "the backup on the replica, retention on the primary")
	assert.Equal(t, []string{"db-1"}, managed.prunedApps)

	// A replica that is gone: the backup is taken on the primary.
	exec.apps = nil
	delete(run.refObjects.RefApps, "db-1-r1")
	_, err = e.runJob(context.Background(), database.Tx{}, run)
	assert.NoError(t, err)
	assert.Equal(t, []string{"db-1", "db-1"}, exec.apps)
}

// In a promoted cluster, retention runs on the member promoted, not on the
// fenced database the job belongs to; pruning is still the database's setting's.
func TestANightlyRunRetainsOnThePromotedMember(t *testing.T) {
	exec := &nightlyExec{outputs: map[string]map[string]string{
		"backup": {"BACKUP_NAME": "base_000000000000000000000002", "BACKUPS": keptBackups(2)},
	}}
	managed := &fakeManaged{}
	e := &Executor{schedJobExecService: exec, managedService: managed}
	run := nightlyRun()
	job := run.jobSetting.MustAsSchedJob()
	job.App = entity.ObjectID{ID: "db-1-r1"}
	run.jobSetting.MustSetData(job)
	run.refObjects.RefApps["db-1-r1"] = &entity.App{ID: "db-1-r1", LogicalParentID: "db-1"}
	setting := run.refObjects.RefSettings["m1"]
	data := setting.MustAsAppManagedSettings()
	data.Postgres.Writer = entity.ObjectID{ID: "db-1-r1"}
	setting.MustSetData(data)

	_, err := e.runJob(context.Background(), database.Tx{}, run)

	assert.NoError(t, err)
	assert.Equal(t, []string{"db-1-r1", "db-1-r1"}, exec.apps)
	assert.Equal(t, []string{"db-1"}, managed.prunedApps)
}
