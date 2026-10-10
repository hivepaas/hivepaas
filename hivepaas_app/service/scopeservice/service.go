package scopeservice

import (
	"context"

	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/infra/database"
)

type Service interface {
	LoadObjectScope(ctx context.Context, db database.IDB, scopeType base.ObjectScopeType,
		objectID string, requireActive bool) (*entity.ObjectScope, error)
	LoadObjectScopeData(ctx context.Context, db database.IDB, scope *entity.ObjectScope) error
	// FilterScope is the scope a list reached through scope shows once filtered to
	// a project, an env or an app (see entity.ObjectScope.FilteredBy): nil when the
	// filter reaches outside the scope, or names an app there is none of.
	FilterScope(ctx context.Context, db database.IDB, scope *entity.ObjectScope,
		projectID, projectEnvID, appID string) (*entity.ObjectScope, error)
}
