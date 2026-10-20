package mariadb

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/managedservice/managedmodel"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/managedservice/walgkit"
)

func promoteReq() *managedmodel.PromoteReq {
	return &managedmodel.PromoteReq{
		Primary: &entity.App{ID: "app-1", Key: "shop-db", Name: "shop-db"},
		Managed: &entity.AppManagedSettings{Engine: "mariadb", Database: &entity.ManagedDatabase{
			Members: []*entity.ManagedMember{{App: entity.ObjectID{ID: "r1"}, Key: "shop-db-r1",
				Role: entity.ManagedRoleReplica, ServerID: 2}}}},
		Member: &entity.ManagedMember{App: entity.ObjectID{ID: "r1"}, Key: "shop-db-r1",
			Role: entity.ManagedRoleReplica, ServerID: 2},
		MountPaths: []string{walgkit.ScriptDir + "/walg-replication", walgkit.ScriptDir + "/walg-entrypoint"},
	}
}

// A promotion fences what writes now, runs walg-entrypoint promote on the
// member - the template has no room for a walg-promote - and names the
// variables that make a standby: not the member's server id and log name, which
// it keeps, its logs going on under their name from the cluster's GTID state.
func TestAPromotionsSpec(t *testing.T) {
	spec, err := New().Promote(promoteReq())

	assert.NoError(t, err)
	assert.Equal(t, [][2]string{{walgkit.EnvFenced, "shop-db-r1 was promoted"}}, spec.FenceEnv)
	assert.Equal(t, walgkit.ScriptDir+"/walg-entrypoint promote", spec.Promote.Command)
	assert.Equal(t, walgkit.ScriptDir+"/walg-replication", spec.Allow.Command, "the replicas let into it")
	assert.Equal(t, []string{walgkit.EnvStandbyHost, walgkit.EnvStandbyName, walgkit.EnvStandbyPasswordFile},
		spec.StandbyEnvKeys)
	assert.NotContains(t, spec.StandbyEnvKeys, envServerID)
	assert.NotContains(t, spec.StandbyEnvKeys, envLogBin)
	assert.Equal(t, walgkit.EnvStandbyHost, spec.StandbyHostKey)
}

func TestAPromotionIsRefused(t *testing.T) {
	old := promoteReq()
	old.MountPaths = []string{walgkit.ScriptDir + "/walg-entrypoint"}
	primary := promoteReq()
	primary.Member.Role = entity.ManagedRolePrimary
	for name, tc := range map[string]struct {
		req  *managedmodel.PromoteReq
		want error
	}{
		"scripts from before promotion": {old, hperrors.ErrAppManagedPromoteUnsupported},
		"a member that is no replica":   {primary, hperrors.ErrArgumentInvalid},
	} {
		_, err := New().Promote(tc.req)

		assert.ErrorIs(t, err, tc.want, name)
	}
}

// The member promoted writes: the engine's jobs run on it, and the database it
// was promoted over is fenced.
func TestReconfigurePromotesAWriter(t *testing.T) {
	req := configureReq(t, deploymentSetting())
	configured, err := New().Configure(req)
	if !assert.NoError(t, err) {
		return
	}
	setting := configured.Settings[0]
	managed := setting.MustAsAppManagedSettings()
	jobs := map[string]*entity.Setting{}
	for _, job := range configured.Settings[1:] {
		jobs[job.ID] = job
	}
	db := managed.Database
	db.Members = []*entity.ManagedMember{
		{App: entity.ObjectID{ID: "r1"}, Key: "shop-db-r1", Role: entity.ManagedRoleReplica, ServerID: 2},
		{App: entity.ObjectID{ID: "r2"}, Key: "shop-db-r2", Role: entity.ManagedRoleReplica, ServerID: 3},
	}

	_, err = New().Reconfigure(&managedmodel.ReconfigureReq{App: &entity.App{ID: "app-1", Key: "shop-db"},
		ManagedID: setting.ID, Managed: managed, Jobs: jobs, Now: time.Now().UTC(),
		Change: &managedmodel.SettingsChange{Writer: new("r1")}})

	assert.NoError(t, err)
	assert.Equal(t, "r1", db.Writer.ID)
	assert.False(t, db.FencedAt.IsZero(), "the database promoted over is fenced")
	assert.Equal(t, entity.ManagedRolePrimary, db.Member("r1").Role)
	for name, id := range map[string]string{"nightly": db.Jobs.Nightly.ID, "health": db.Jobs.Health.ID,
		"restore test": db.Jobs.RestoreTest.ID} {
		assert.Equal(t, "r1", jobs[id].MustAsSchedJob().App.ID, name)
	}
}
