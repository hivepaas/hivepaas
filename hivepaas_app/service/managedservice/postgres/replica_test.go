package postgres

import (
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/managedservice/managedmodel"
)

func replicaReq(placement string) *managedmodel.ReplicaReq {
	return &managedmodel.ReplicaReq{
		Primary: &entity.App{ID: "app-1", Key: "shop-db", Name: "shop-db"},
		Managed: &entity.AppManagedSettings{Engine: "postgres", EngineAPI: 1,
			Locks:    entity.ManagedLocks{Mounts: []string{volumeArchiveDir, "/var/lib/postgresql"}},
			Database: &entity.ManagedDatabase{Storage: entity.ManagedStorage{Placement: placement}}},
		Key:        "shop-db-r1",
		MountPaths: []string{scriptDir + "/walg-entrypoint", scriptDir + "/walg-replication", secretsDir + "/x"},
	}
}

// A replica streams from the primary by its key, is known there by its own, and
// reads the replication password from a file; its data goes where the
// primary's does. The primary lets it in with walg-replication.
func TestAReplicasSpec(t *testing.T) {
	spec, err := New().Replica(replicaReq(entity.ManagedPlacementOffSite))

	assert.NoError(t, err)
	assert.Equal(t, [][2]string{
		{envStandbyHost, "shop-db"},
		{envStandbyName, "shop-db-r1"},
		{envStandbyPasswordFile, replicaPasswordFile},
	}, spec.Env)
	assert.Equal(t, replicaPasswordFile, spec.PasswordFile)
	assert.True(t, strings.HasPrefix(spec.PasswordFile, "/run/secrets/"))
	assert.NotEmpty(t, spec.SecretName)
	assert.Equal(t, "/var/lib/postgresql", spec.DataTarget, "not the archive's volume")
	assert.Equal(t, scriptDir+"/walg-replication", spec.Allow.Command)
}

func TestAReplicaIsRefused(t *testing.T) {
	old := replicaReq(entity.ManagedPlacementOffSite)
	old.MountPaths = []string{scriptDir + "/walg-entrypoint"}
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

// The primary's health check lists the replicas streaming from it; a member it
// does not list is disconnected, and one no check has looked at is unknown.
func TestStatusOfTheMembers(t *testing.T) {
	managed := statusManaged(entity.ManagedRestoreTestContainer)
	at := time.Date(2026, 10, 19, 9, 0, 0, 0, time.UTC)
	managed.Database.Members = []*entity.ManagedMember{
		{App: entity.ObjectID{ID: "r1"}, Key: "db-r1", Role: entity.ManagedRoleReplica, Node: "n1", At: at},
		{App: entity.ObjectID{ID: "r2"}, Key: "db-r2", Role: entity.ManagedRoleReplica, Node: "n2", At: at},
	}
	runs := healthyRuns()
	runs["health"].Outputs[outReplicas] = "db-r1@streaming@8192@0.4 other@catchup@1@"

	status, err := New().Status(&managedmodel.StatusReq{Managed: managed, Runs: runs, Now: statusNow})

	assert.NoError(t, err)
	if assert.Len(t, status.Members, 2) {
		first, second := status.Members[0], status.Members[1]
		assert.Equal(t, "r1", first.Member.App.ID)
		assert.Equal(t, "streaming", first.State)
		assert.Equal(t, int64(8192), first.LagBytes)
		if assert.NotNil(t, first.LagSeconds) {
			assert.InDelta(t, 0.4, *first.LagSeconds, 0.001)
		}
		assert.NotNil(t, first.CheckedAt)
		assert.Equal(t, managedmodel.MemberDisconnected, second.State)
	}

	delete(runs, "health")
	status, err = New().Status(&managedmodel.StatusReq{Managed: managed, Runs: runs, Now: statusNow})
	assert.NoError(t, err)
	assert.Equal(t, managedmodel.MemberUnknown, status.Members[0].State)
	assert.Nil(t, status.Members[0].LagSeconds)

	// A health script from before replicas prints no list: nothing is known.
	runs = healthyRuns()
	status, err = New().Status(&managedmodel.StatusReq{Managed: managed, Runs: runs, Now: statusNow})
	assert.NoError(t, err)
	assert.Equal(t, managedmodel.MemberUnknown, status.Members[0].State)
}
