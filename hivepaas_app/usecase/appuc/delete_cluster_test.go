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
// promoted over holds the cluster, deleted only with all of it, as asked; any
// other app is deleted as ever.
func TestAClusterIsNotDeletedInPart(t *testing.T) {
	member := managedApp("db-r1", &entity.AppManagedSettings{Engine: "postgres",
		MemberOf: &entity.ManagedMemberOf{App: entity.ObjectID{ID: "db"}, Role: entity.ManagedRoleReplica}})
	assert.ErrorIs(t, refuseClusterDeletion(member, false), hperrors.ErrAppManagedIsMember)
	assert.ErrorIs(t, refuseClusterDeletion(member, true), hperrors.ErrAppManagedIsMember,
		"a member is never the cluster")

	held := managedApp("db", &entity.AppManagedSettings{Engine: "postgres",
		Postgres: &entity.ManagedPostgres{Writer: entity.ObjectID{ID: "db-r1"}}})
	assert.ErrorIs(t, refuseClusterDeletion(held, false), hperrors.ErrAppManagedClusterHeld)
	assert.NoError(t, refuseClusterDeletion(held, true), "the whole cluster, as asked")

	primary := managedApp("db", &entity.AppManagedSettings{Engine: "postgres", Postgres: &entity.ManagedPostgres{}})
	assert.NoError(t, refuseClusterDeletion(primary, false))
	assert.NoError(t, refuseClusterDeletion(&entity.App{ID: "web"}, false))
}
