package permission

import (
	"context"

	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/basedto"
	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/infra/database"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/bunex"
)

type Manager interface {
	CheckAccess(ctx context.Context, db database.IDB, auth *basedto.Auth, check AccessCheck) (bool, error)

	// NOTE: this func should be called within a transaction
	UpdateACLPermissions(ctx context.Context, db database.IDB, perms []*entity.ACLPermission) error
	RemoveACLPermissions(ctx context.Context, db database.IDB, perms []*base.PermissionResource) error
	RemoveACLPermissionsBySubjects(ctx context.Context, db database.IDB,
		subjectType base.SubjectType, subjectIDs []string) error
	RemoveACLPermissionsByObjects(ctx context.Context, db database.IDB, objectIDs []string) error
	RemoveACLPermissionsOfUsers(ctx context.Context, db database.IDB, userIDs []string) error

	LoadProjectAccesses(ctx context.Context, db database.IDB, projectID string, projectEnvIDs []string,
		makeAdjustment bool, extraLoadOpts ...bunex.SelectQueryOption) (modPerms []*entity.ACLPermission,
		projectPerms []*entity.ACLPermission, envPerms map[string][]*entity.ACLPermission, err error)
	LoadProjectAccessUsers(ctx context.Context, db database.IDB, projectID string, projectEnvIDs []string,
		extraLoadOpts ...bunex.SelectQueryOption) (userPerms []*entity.ACLPermission, err error)
}
