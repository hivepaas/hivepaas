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

func replicaReq(placement string) *managedmodel.ReplicaReq {
	return &managedmodel.ReplicaReq{
		Primary: &entity.App{ID: "app-1", Key: "shop-db", Name: "shop-db"},
		Managed: &entity.AppManagedSettings{Engine: "mariadb", EngineAPI: 1,
			Locks:    entity.ManagedLocks{Mounts: []string{walgkit.VolumeArchiveDir, "/var/lib/mysql"}},
			Database: &entity.ManagedDatabase{Storage: entity.ManagedStorage{Placement: placement}}},
		Key: "shop-db-r1",
		MountPaths: []string{walgkit.ScriptDir + "/walg-entrypoint", walgkit.ScriptDir + "/walg-replication",
			walgkit.SecretsDir + "/x"},
	}
}

// A replica streams from the writer by its key, is listed there by its own,
// reads the replication password from a file the mysql user owns, and has a
// server id and binary logs of its own: the logs of a member promoted go into
// the cluster's archive beside the others'. Its data goes where the primary's
// does; the primary lets it in with walg-replication.
func TestAReplicasSpec(t *testing.T) {
	spec, err := New().Replica(replicaReq(entity.ManagedPlacementOffSite))

	assert.NoError(t, err)
	assert.Equal(t, [][2]string{
		{walgkit.EnvStandbyHost, "shop-db"},
		{walgkit.EnvStandbyName, "shop-db-r1"},
		{walgkit.EnvStandbyPasswordFile, walgkit.ReplicaPasswordFile},
		{envServerID, "2"},
		{envLogBin, "mariadb-bin-shop-db-r1"},
	}, spec.Env)
	assert.Equal(t, 2, spec.ServerID)
	assert.Equal(t, mysqlUID, spec.FileOwner)
	assert.Equal(t, "/var/lib/mysql", spec.DataTarget, "not the archive's volume")
	assert.Equal(t, walgkit.ScriptDir+"/walg-replication", spec.Allow.Command)
}

// Each member's server id is the first no member has, from 2: the cluster's
// app's is 1, in its command.
func TestAReplicaTakesTheFirstFreeServerID(t *testing.T) {
	req := replicaReq(entity.ManagedPlacementOffSite)
	req.Managed.Database.Members = []*entity.ManagedMember{
		{Key: "shop-db-r1", Role: entity.ManagedRoleReplica, ServerID: 2},
		{Key: "shop-db-r3", Role: entity.ManagedRoleReplica, ServerID: 4},
	}
	req.Key = "shop-db-r2"

	spec, err := New().Replica(req)

	assert.NoError(t, err)
	assert.Equal(t, 3, spec.ServerID)
	assert.Contains(t, spec.Env, [2]string{envServerID, "3"})
	assert.Contains(t, spec.Env, [2]string{envLogBin, "mariadb-bin-shop-db-r2"})
}

func TestAReplicaIsRefused(t *testing.T) {
	old := replicaReq(entity.ManagedPlacementOffSite)
	old.MountPaths = []string{walgkit.ScriptDir + "/walg-entrypoint"}
	restoring := replicaReq(entity.ManagedPlacementOffSite)
	restoring.Managed.Database.Restore = &entity.ManagedRestore{SourceApp: entity.ObjectID{ID: "src"}}
	member := replicaReq(entity.ManagedPlacementOffSite)
	member.Managed.MemberOf = &entity.ManagedMemberOf{App: entity.ObjectID{ID: "p"}, Role: entity.ManagedRoleReplica}
	for name, tc := range map[string]struct {
		req  *managedmodel.ReplicaReq
		want error
	}{
		"scripts from before replicas": {old, hperrors.ErrAppManagedReplicaUnsupported},
		"an archive on a volume": {replicaReq(entity.ManagedPlacementShared),
			hperrors.ErrAppManagedReplicaNeedsObjectStorage},
		"a database still restoring": {restoring, hperrors.ErrAppManagedSwitchNotReady},
		"a replica itself":           {member, hperrors.ErrAppManagedIsMember},
	} {
		_, err := New().Replica(tc.req)

		assert.ErrorIs(t, err, tc.want, name)
	}
}

// The writer's health check lists the replicas streaming from it, each with
// what it has not applied; a member it does not list is disconnected.
func TestStatusOfTheMembers(t *testing.T) {
	managed := statusManaged()
	at := time.Date(2026, 10, 19, 9, 0, 0, 0, time.UTC)
	managed.Database.Members = []*entity.ManagedMember{
		{App: entity.ObjectID{ID: "r1"}, Key: "db-r1", Role: entity.ManagedRoleReplica, ServerID: 2, At: at},
		{App: entity.ObjectID{ID: "r2"}, Key: "db-r2", Role: entity.ManagedRoleReplica, ServerID: 3, At: at},
	}
	runs := map[string]*managedmodel.JobRun{"health": {EndedAt: statusNow.Add(-time.Minute),
		Outputs: map[string]string{walgkit.OutReplicas: "db-r1@streaming@4096@2"}}}

	status, err := New().Status(&managedmodel.StatusReq{Managed: managed, Runs: runs, Now: statusNow})

	assert.NoError(t, err)
	if assert.Len(t, status.Members, 2) {
		assert.Equal(t, "streaming", status.Members[0].State)
		assert.Equal(t, int64(4096), status.Members[0].LagBytes)
		if assert.NotNil(t, status.Members[0].LagSeconds) {
			assert.InDelta(t, 2.0, *status.Members[0].LagSeconds, 0.001)
		}
		assert.Equal(t, managedmodel.MemberDisconnected, status.Members[1].State)
	}
}
