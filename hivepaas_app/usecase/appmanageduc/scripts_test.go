package appmanageduc

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/managedservice/managedscripts"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/appmanageduc/appmanageddto"
)

// A cluster's scripts run on its members and on its verifier too.
func TestClusterMembers(t *testing.T) {
	managed := &entity.AppManagedSettings{Database: &entity.ManagedDatabase{
		Members: []*entity.ManagedMember{{App: entity.ObjectID{ID: "r1"}, Key: "db-r1"}},
		RestoreTest: entity.ManagedRestoreTest{Mode: entity.ManagedRestoreTestVerifier,
			VerifierApp: entity.ObjectID{ID: "v1"}},
	}}

	members := clusterMembers(managed)

	if assert.Len(t, members, 2) {
		assert.Equal(t, "r1", members[0].App.ID)
		assert.Equal(t, "v1", members[1].App.ID)
		assert.Equal(t, "verifier", members[1].Key)
	}
	assert.Empty(t, clusterMembers(&entity.AppManagedSettings{}))
}

// The member that writes is deployed last; the others while it serves.
func TestDeployOrder(t *testing.T) {
	db, r1, v := &entity.App{ID: "db"}, &entity.App{ID: "r1"}, &entity.App{ID: "v"}

	assert.Equal(t, []*entity.App{r1, v, db}, deployOrder([]*entity.App{db, r1, v}, "db"))
	assert.Equal(t, []*entity.App{db, v, r1}, deployOrder([]*entity.App{db, r1, v}, "r1"), "promoted")
}

// An upgrade applies what was reviewed, and only when there is something to
// apply: a member out of step is.
func TestCheckUpgradable(t *testing.T) {
	available := &managedscripts.Set{Scripts: []*managedscripts.Script{{Path: "/opt/hivepaas/bin/walg-lib"}}}
	view := func(state string, outOfStep int) *appmanageddto.ManagedScriptsView {
		v := &appmanageddto.ManagedScriptsView{Offer: &managedscripts.Offer{State: state, Available: available}}
		for range outOfStep {
			v.OutOfStep = append(v.OutOfStep, &appmanageddto.ManagedScriptsMemberResp{Key: "db-r1"})
		}
		return v
	}
	fp := available.Fingerprint()

	assert.NoError(t, checkUpgradable(view(managedscripts.StateUpdateAvailable, 0), fp))
	assert.NoError(t, checkUpgradable(view(managedscripts.StateUpToDate, 1), fp), "a member out of step")
	assert.ErrorIs(t, checkUpgradable(view(managedscripts.StateUpdateAvailable, 0), "other"),
		hperrors.ErrUpdateVerMismatched)
	for _, state := range []string{managedscripts.StateUpToDate, managedscripts.StateIncompatible,
		managedscripts.StateUnavailable} {
		assert.ErrorIs(t, checkUpgradable(view(state, 0), fp), hperrors.ErrArgumentInvalid, state)
	}
}
