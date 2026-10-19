package postgres

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/managedservice/managedmodel"
)

func promoteReq() *managedmodel.PromoteReq {
	return &managedmodel.PromoteReq{
		Primary: &entity.App{ID: "app-1", Key: "shop-db", Name: "shop-db"},
		Managed: &entity.AppManagedSettings{Engine: "postgres", Database: &entity.ManagedDatabase{
			Members: []*entity.ManagedMember{{App: entity.ObjectID{ID: "r1"}, Key: "shop-db-r1",
				Role: entity.ManagedRoleReplica}}}},
		Member:     &entity.ManagedMember{App: entity.ObjectID{ID: "r1"}, Key: "shop-db-r1", Role: entity.ManagedRoleReplica},
		MountPaths: []string{scriptDir + "/walg-replication", scriptDir + "/walg-promote"},
	}
}

// A promotion fences what writes now by a variable walg-entrypoint reads, runs
// walg-promote on the member, and names the variables that make a standby.
func TestAPromotionsSpec(t *testing.T) {
	spec, err := New().Promote(promoteReq())

	assert.NoError(t, err)
	assert.Equal(t, [][2]string{{envFenced, "shop-db-r1 was promoted"}}, spec.FenceEnv)
	assert.Equal(t, scriptDir+"/walg-promote", spec.Promote.Command)
	assert.Equal(t, scriptDir+"/walg-replication", spec.Allow.Command, "the replicas let into it")
	assert.Equal(t, []string{envStandbyHost, envStandbyName, envStandbyPasswordFile}, spec.StandbyEnvKeys)
	assert.Equal(t, envStandbyHost, spec.StandbyHostKey)
}

func TestAPromotionIsRefused(t *testing.T) {
	old := promoteReq()
	old.MountPaths = []string{scriptDir + "/walg-replication"}
	primary := promoteReq()
	primary.Member.Role = entity.ManagedRolePrimary
	member := promoteReq()
	member.Managed.MemberOf = &entity.ManagedMemberOf{App: entity.ObjectID{ID: "p"}, Role: entity.ManagedRoleReplica}
	restoring := promoteReq()
	restoring.Managed.Database.Restore = &entity.ManagedRestore{SourceApp: entity.ObjectID{ID: "src"}}
	for name, tc := range map[string]struct {
		req  *managedmodel.PromoteReq
		want error
	}{
		"scripts from before promotion": {old, hperrors.ErrAppManagedPromoteUnsupported},
		"a member that is no replica":   {primary, hperrors.ErrArgumentInvalid},
		"on a member":                   {member, hperrors.ErrAppManagedIsMember},
		"a database still restoring":    {restoring, hperrors.ErrAppManagedSwitchNotReady},
	} {
		_, err := New().Promote(tc.req)

		assert.ErrorIs(t, err, tc.want, name)
	}
}

// promoted reconfigures the writer to the member given.
func promoted(t *testing.T, setting *entity.Setting, managed *entity.AppManagedSettings,
	jobs map[string]*entity.Setting, member string) *managedmodel.ReconfigureResp {
	t.Helper()
	resp, err := New().Reconfigure(&managedmodel.ReconfigureReq{App: &entity.App{ID: "app-1", Key: "shop-db"},
		ManagedID: setting.ID, Managed: managed, Jobs: jobs, Now: time.Now().UTC(),
		Change: &managedmodel.SettingsChange{Writer: new(member)}})
	if !assert.NoError(t, err) {
		t.FailNow()
	}
	return resp
}

// The member promoted writes: the engine's jobs run on it, the base backups too
// unless another replica takes them, and the database it was promoted over is
// fenced. A second promotion fences the member that wrote.
func TestReconfigurePromotesAWriter(t *testing.T) {
	setting, managed, jobs := configured(t, &managedmodel.Options{DumpRepo: "repo-1"})
	pg := managed.Database
	pg.Members = []*entity.ManagedMember{
		{App: entity.ObjectID{ID: "r1"}, Key: "shop-db-r1", Role: entity.ManagedRoleReplica},
		{App: entity.ObjectID{ID: "r2"}, Key: "shop-db-r2", Role: entity.ManagedRoleReplica},
	}
	pg.BackupFrom = entity.ObjectID{ID: "r1"}
	appOf := func(id string) string { return jobs[id].MustAsSchedJob().App.ID }

	resp := promoted(t, setting, managed, jobs, "r1")

	assert.Equal(t, "r1", pg.Writer.ID)
	assert.False(t, pg.FencedAt.IsZero(), "the database promoted over is fenced")
	assert.Equal(t, entity.ManagedRolePrimary, pg.Member("r1").Role)
	assert.Equal(t, entity.ManagedRoleReplica, pg.Member("r2").Role)
	assert.Empty(t, pg.BackupFrom.ID, "the writer takes them itself")
	for name, id := range map[string]string{"nightly": pg.Jobs.Nightly.ID, "health": pg.Jobs.Health.ID,
		"restore test": pg.Jobs.RestoreTest.ID, "dump": pg.Jobs.Dump.ID} {
		assert.Equal(t, "r1", appOf(id), name)
	}
	assert.Len(t, resp.Changed, 4)
	assert.Empty(t, resp.Rescheduled, "when they run stays")
	assert.Equal(t, "shop-db-r1", pg.WriterKey(&entity.App{Key: "shop-db"}))

	// The backups on r2, then a promotion of r2: r1 is fenced, the jobs move.
	pg.BackupFrom = entity.ObjectID{ID: "r2"}
	promoted(t, setting, managed, jobs, "r2")
	assert.Equal(t, entity.ManagedRoleFenced, pg.Member("r1").Role)
	assert.Equal(t, entity.ManagedRolePrimary, pg.Member("r2").Role)
	assert.Equal(t, "r2", appOf(pg.Jobs.Health.ID))
	assert.Equal(t, "r2", appOf(pg.Jobs.Nightly.ID))

	_, err := New().Reconfigure(&managedmodel.ReconfigureReq{App: &entity.App{ID: "app-1"}, Managed: managed,
		Jobs: jobs, Change: &managedmodel.SettingsChange{Writer: new("r1")}})
	assert.ErrorIs(t, err, hperrors.ErrArgumentInvalid, "a fenced member is not promoted")
}

// With a writer, base backups taken on a replica come home to the writer, not to
// the fenced database; a new replica streams from the writer.
func TestTheWriterIsWhereTheDatabaseIs(t *testing.T) {
	setting, managed, jobs := configured(t, nil)
	pg := managed.Database
	pg.Members = []*entity.ManagedMember{
		{App: entity.ObjectID{ID: "r1"}, Key: "shop-db-r1", Role: entity.ManagedRoleReplica},
		{App: entity.ObjectID{ID: "r2"}, Key: "shop-db-r2", Role: entity.ManagedRoleReplica},
	}
	promoted(t, setting, managed, jobs, "r1")
	_, err := New().Reconfigure(&managedmodel.ReconfigureReq{App: &entity.App{ID: "app-1", Key: "shop-db"},
		ManagedID: setting.ID, Managed: managed, Jobs: jobs,
		Change: &managedmodel.SettingsChange{BackupFrom: new("")}})
	assert.NoError(t, err)
	assert.Equal(t, "r1", jobs[pg.Jobs.Nightly.ID].MustAsSchedJob().App.ID)

	req := replicaReq(entity.ManagedPlacementOffSite)
	req.Managed = managed
	managed.Locks.Mounts = []string{"/var/lib/postgresql"}
	pg.Storage.Placement = entity.ManagedPlacementOffSite
	req.Writer = &entity.App{ID: "r1", Key: "shop-db-r1"}
	spec, err := New().Replica(req)
	assert.NoError(t, err)
	assert.Contains(t, spec.Env, [2]string{envStandbyHost, "shop-db-r1"})
}

func TestStatusOfAPromotedCluster(t *testing.T) {
	managed := statusManaged(entity.ManagedRestoreTestContainer)
	managed.Database.Members = []*entity.ManagedMember{
		{App: entity.ObjectID{ID: "r1"}, Key: "db-r1", Role: entity.ManagedRolePrimary},
		{App: entity.ObjectID{ID: "r2"}, Key: "db-r2", Role: entity.ManagedRoleReplica},
		{App: entity.ObjectID{ID: "r3"}, Key: "db-r3", Role: entity.ManagedRoleFenced},
	}
	runs := healthyRuns()
	runs["health"].Outputs[outReplicas] = "db-r2@streaming@0@"

	status, err := New().Status(&managedmodel.StatusReq{Managed: managed, Runs: runs, Now: statusNow})

	assert.NoError(t, err)
	states := map[string]string{}
	for _, member := range status.Members {
		states[member.Member.Key] = member.State
	}
	assert.Equal(t, map[string]string{"db-r1": "primary", "db-r2": "streaming", "db-r3": "fenced"}, states)
}
