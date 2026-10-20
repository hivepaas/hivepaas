package mariadb

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/managedservice/managedmodel"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/managedservice/walgkit"
)

// The verifier runs no server of its own: walg-entrypoint hands sleep to the
// image's entrypoint, and walg-restore-test starts the copy it restores into.
func TestTheVerifierRunsNoServer(t *testing.T) {
	spec := New().Verifier()

	assert.Equal(t, "sleep infinity", spec.Command)
	assert.Equal(t, walgkit.VerifierScratchDir, spec.ScratchDir)
}

// The test reaches the live database over the project's network for its point,
// as root, and restores into the verifier's volume.
func TestTheRestoreTestMovesIntoAVerifier(t *testing.T) {
	managed, jobs := reconfigureSetup(t)

	resp, err := New().Reconfigure(&managedmodel.ReconfigureReq{App: &entity.App{ID: "app-1", Key: "shop-db"},
		Managed: managed, Jobs: jobs, Change: &managedmodel.SettingsChange{
			RestoreTest: &managedmodel.RestoreTestChange{Verifier: &entity.App{ID: "ver-1"}, NodeID: "node-2"}}})

	assert.NoError(t, err)
	assert.Equal(t, entity.ManagedRestoreTestVerifier, managed.Database.RestoreTest.Mode)
	setting := jobs[managed.Database.Jobs.RestoreTest.ID]
	job := setting.MustAsSchedJob()
	assert.Equal(t, base.SchedJobTypeManagedRestoreTest, job.JobType)
	assert.Equal(t, "ver-1", job.App.ID)
	env := map[string]string{}
	for _, v := range job.Command.EnvVars {
		env[v.Key] = v.Value
	}
	assert.Equal(t, "shop-db", env[walgkit.EnvRestoreSourceHost])
	assert.Equal(t, walgkit.VerifierScratchDir, env[walgkit.EnvRestoreScratchDir])
	assert.Len(t, resp.Changed, 1)
}
