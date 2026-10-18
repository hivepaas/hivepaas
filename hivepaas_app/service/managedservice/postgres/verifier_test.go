package postgres

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/managedservice/managedmodel"
)

var shopDB = &entity.App{ID: "app-1", Key: "shop-db"}

func reconfigureRestoreTest(t *testing.T, managed *entity.AppManagedSettings, jobs map[string]*entity.Setting,
	change *managedmodel.RestoreTestChange) *managedmodel.ReconfigureResp {
	t.Helper()
	resp, err := New().Reconfigure(&managedmodel.ReconfigureReq{App: shopDB, Managed: managed, Jobs: jobs,
		Change: &managedmodel.SettingsChange{RestoreTest: change}})
	assert.NoError(t, err)
	return resp
}

// The verifier has no database of its own: the test reaches the live one over
// the project's network for its restore point, and restores into its volume.
func TestTheRestoreTestMovesIntoAVerifier(t *testing.T) {
	managed, jobs := reconfigureSetup(t)

	resp := reconfigureRestoreTest(t, managed, jobs, &managedmodel.RestoreTestChange{
		Verifier: &entity.App{ID: "ver-1"}, NodeID: "node-2", KeyAuth: entity.ObjectID{ID: "key-ro"}})

	assert.Equal(t, entity.ManagedRestoreTest{Mode: entity.ManagedRestoreTestVerifier,
		VerifierApp: entity.ObjectID{ID: "ver-1"}, NodeID: "node-2", KeyAuth: entity.ObjectID{ID: "key-ro"}},
		managed.Postgres.RestoreTest)
	setting := jobs[managed.Postgres.Jobs.RestoreTest.ID]
	job := setting.MustAsSchedJob()
	assert.Equal(t, base.SchedJobTypeManagedRestoreTest, job.JobType)
	assert.Equal(t, string(base.SchedJobTypeManagedRestoreTest), setting.Kind)
	assert.Equal(t, "ver-1", job.App.ID)
	assert.Equal(t, "shop-db", jobEnv(job, envRestoreSourceHost))
	assert.Equal(t, verifierScratchDir, jobEnv(job, envRestoreScratchDir))
	assert.Equal(t, scriptDir+"/walg-restore-test", job.Command.Command, "the same script")
	assert.Equal(t, []string{setting.ID}, ids(resp.Changed))
	assert.Empty(t, resp.Rescheduled, "it runs when it did")
}

func TestTheRestoreTestMovesBackIntoTheContainer(t *testing.T) {
	managed, jobs := reconfigureSetup(t)
	managed.Postgres.RestoreTest.CheckQuery = "SELECT 1"
	reconfigureRestoreTest(t, managed, jobs, &managedmodel.RestoreTestChange{Verifier: &entity.App{ID: "ver-1"}})

	resp := reconfigureRestoreTest(t, managed, jobs, &managedmodel.RestoreTestChange{})

	assert.Equal(t, entity.ManagedRestoreTest{Mode: entity.ManagedRestoreTestContainer, CheckQuery: "SELECT 1"},
		managed.Postgres.RestoreTest)
	setting := jobs[managed.Postgres.Jobs.RestoreTest.ID]
	job := setting.MustAsSchedJob()
	assert.Equal(t, base.SchedJobTypeContainerCommand, job.JobType)
	assert.Equal(t, string(base.SchedJobTypeContainerCommand), setting.Kind)
	assert.Equal(t, "app-1", job.App.ID)
	for _, env := range job.Command.EnvVars {
		assert.NotContains(t, []string{envRestoreSourceHost, envRestoreScratchDir}, env.Key)
	}
	assert.Len(t, resp.Changed, 1)
}

// The verifier's command is not the database's: nothing in it runs a server, so
// nothing in it archives - the server the test starts has archiving off.
func TestTheVerifierRunsNoServer(t *testing.T) {
	spec := New().Verifier()

	assert.Equal(t, "sleep infinity", spec.Command)
	assert.NotContains(t, spec.Command, "postgres")
	assert.Equal(t, verifierScratchDir, spec.ScratchDir)
}
