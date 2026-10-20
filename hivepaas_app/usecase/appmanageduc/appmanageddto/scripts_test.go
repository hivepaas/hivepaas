package appmanageddto

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/hivepaas/hivepaas/hivepaas_app/service/managedservice/managedscripts"
)

// An offer of nothing has no available revision; lists are never null.
func TestTransformManagedScripts(t *testing.T) {
	current := &managedscripts.Set{Scripts: []*managedscripts.Script{{Path: "/opt/hivepaas/bin/walg-lib"}}}
	resp := TransformManagedScripts(&ManagedScriptsView{Current: current, CurrentRevision: "abc",
		Offer: &managedscripts.Offer{State: managedscripts.StateUnavailable, Reason: "gone"}})

	assert.Equal(t, "unavailable", resp.State)
	assert.Equal(t, "gone", resp.Reason)
	assert.Equal(t, &ManagedScriptsRevisionResp{Revision: "abc", Fingerprint: current.Fingerprint()}, resp.Current)
	assert.Nil(t, resp.Available)
	assert.NotNil(t, resp.Changes)
	assert.NotNil(t, resp.OutOfStep)

	available := &managedscripts.Set{}
	resp = TransformManagedScripts(&ManagedScriptsView{Current: current, AvailableRevision: "def",
		Offer: &managedscripts.Offer{State: managedscripts.StateUpdateAvailable, Available: available,
			Changes: []*managedscripts.Change{{Kind: "artifact", Name: "wal-g", Change: "changed", From: "1", To: "2"}}}})
	assert.Equal(t, &ManagedScriptsRevisionResp{Revision: "def", Fingerprint: available.Fingerprint()}, resp.Available)
	assert.Equal(t, []*ManagedScriptChangeResp{{Kind: "artifact", Name: "wal-g", Change: "changed", From: "1", To: "2"}},
		resp.Changes)
}
