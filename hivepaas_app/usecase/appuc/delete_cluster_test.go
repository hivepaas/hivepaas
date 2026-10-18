package appuc

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
)

func managedApp(name string, managed *entity.AppManagedSettings) *entity.App {
	setting := &entity.Setting{ID: "m-" + name, Type: base.SettingTypeAppManaged, Status: base.SettingStatusActive}
	setting.MustSetData(managed)
	return &entity.App{ID: name, Name: name, Settings: []*entity.Setting{setting}}
}

// A cluster's member is removed from its primary; the app a replica was
// promoted over holds the cluster; any other app is deleted as ever.
func TestAClusterIsNotDeletedInPart(t *testing.T) {
	member := managedApp("db-r1", &entity.AppManagedSettings{Engine: "postgres",
		MemberOf: &entity.ManagedMemberOf{App: entity.ObjectID{ID: "db"}, Role: entity.ManagedRoleReplica}})
	assert.ErrorIs(t, refuseClusterDeletion(member), hperrors.ErrAppManagedIsMember)

	held := managedApp("db", &entity.AppManagedSettings{Engine: "postgres",
		Postgres: &entity.ManagedPostgres{Writer: entity.ObjectID{ID: "db-r1"}}})
	assert.ErrorIs(t, refuseClusterDeletion(held), hperrors.ErrAppManagedClusterHeld)

	primary := managedApp("db", &entity.AppManagedSettings{Engine: "postgres", Postgres: &entity.ManagedPostgres{}})
	assert.NoError(t, refuseClusterDeletion(primary))
	assert.NoError(t, refuseClusterDeletion(&entity.App{ID: "web"}))
}
