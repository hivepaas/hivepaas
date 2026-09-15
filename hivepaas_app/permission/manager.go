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

	// Check access on a specific setting
	CheckAccessOnSetting(ctx context.Context, db database.IDB, auth *basedto.Auth, check AccessCheck,
		setting *entity.Setting) (bool, error)

	// AuthorizeAccessChanges checks the access changes the acting user wants to
	// apply to another subject - a non-admin may only flip the action bits they
	// hold themselves, both when granting and when revoking - and returns the
	// subject's existing rows the operation may replace.
	AuthorizeAccessChanges(ctx context.Context, db database.IDB, auth *basedto.Auth,
		desired []*entity.ACLPermission, current []*entity.ACLPermission) ([]*entity.ACLPermission, error)

	// HasCapability reports whether the caller holds a capability
	HasCapability(ctx context.Context, db database.IDB, auth *basedto.Auth,
		capability base.ResourceCapability) (bool, error)

	// AuthorizeSecretReveal decides whether the caller may see a secret in the clear,
	// and records the answer either way.
	AuthorizeSecretReveal(ctx context.Context, db database.IDB, auth *basedto.Auth, subject *RevealSubject) error

	// NOTE: this func should be called within a transaction
	UpdateACLPermissions(ctx context.Context, db database.IDB, perms []*entity.ACLPermission) error
	DeleteACLPermissions(ctx context.Context, db database.IDB, perms []*base.PermissionResource) error
	DeleteACLPermissionsBySubjects(ctx context.Context, db database.IDB,
		subjectType base.SubjectType, subjectIDs []string) error
	DeleteACLPermissionsByObjects(ctx context.Context, db database.IDB, objectIDs []string) error
	DeleteACLPermissionsOfUsers(ctx context.Context, db database.IDB, userIDs []string) error

	// Project permissions
	LoadProjectRawAccesses(ctx context.Context, db database.IDB, projectID string, projectEnvIDs []string,
		extraOpts ...bunex.SelectQueryOption) ([]*entity.ACLPermission, error)
	LoadProjectAccesses(ctx context.Context, db database.IDB, projectID string, projectEnvIDs []string,
		makeAdjustment bool) (modPerms []*entity.ACLPermission, projectPerms []*entity.ACLPermission,
		envPerms map[string][]*entity.ACLPermission, err error)
	LoadProjectAccessUsers(ctx context.Context, db database.IDB, projectID string, projectEnvIDs []string) (
		userPerms []*entity.ACLPermission, err error)
	DeleteProjectAccesses(ctx context.Context, db database.IDB, projectID string,
		extraOpts ...bunex.DeleteQueryOption) error
}
