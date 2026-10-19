package postgres

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/managedservice/managedmodel"
)

// Base backups move to a replica, and back: the nightly job runs where they are
// taken, on the same schedule.
func TestReconfigureMovesTheBaseBackups(t *testing.T) {
	setting, managed, jobs := configured(t, nil)
	managed.Database.Members = []*entity.ManagedMember{{App: entity.ObjectID{ID: "app-1-r1"}, Key: "shop-db-r1",
		Role: entity.ManagedRoleReplica}}
	nightlyID := managed.Database.Jobs.Nightly.ID
	reconfigure := func(backupFrom string) (*managedmodel.ReconfigureResp, error) {
		return New().Reconfigure(&managedmodel.ReconfigureReq{App: &entity.App{ID: "app-1", Key: "shop-db"},
			ManagedID: setting.ID, Managed: managed, Jobs: jobs, Now: time.Now().UTC(),
			Change: &managedmodel.SettingsChange{BackupFrom: new(backupFrom)}})
	}

	resp, err := reconfigure("app-1-r1")

	assert.NoError(t, err)
	assert.Equal(t, "app-1-r1", managed.Database.BackupFrom.ID)
	assert.Equal(t, []string{nightlyID}, ids(resp.Changed))
	assert.Empty(t, resp.Rescheduled, "when it runs stays")
	nightly := jobs[nightlyID].MustAsSchedJob()
	assert.Equal(t, "app-1-r1", nightly.App.ID)
	assert.Equal(t, defaultBaseBackupSchedule, nightly.Schedule.CronExpr)

	resp, err = reconfigure("app-1-r1")
	assert.NoError(t, err)
	assert.Empty(t, resp.Changed, "nothing to move")

	resp, err = reconfigure("")
	assert.NoError(t, err)
	assert.Empty(t, managed.Database.BackupFrom.ID, "the primary")
	assert.Equal(t, []string{nightlyID}, ids(resp.Changed))
	assert.Equal(t, "app-1", jobs[nightlyID].MustAsSchedJob().App.ID)

	_, err = reconfigure("someone-else")
	assert.ErrorIs(t, err, hperrors.ErrArgumentInvalid, "only a replica of this database")
}
