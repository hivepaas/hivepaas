package appmanageduc

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
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
