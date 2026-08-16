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
}
