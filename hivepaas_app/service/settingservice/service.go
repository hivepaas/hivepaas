package settingservice

import (
	"context"

	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/infra/database"
)

type Service interface {
	PersistSettingData(ctx context.Context, db database.IDB, data *PersistingSettingData) error

	LoadRefObjects(ctx context.Context, db database.IDB, refObjects **entity.RefObjects,
		scope *entity.ObjectScope, requireActive bool, inSettings ...*entity.Setting) error
	LoadRefObjectsSkipMissing(ctx context.Context, db database.IDB, refObjects **entity.RefObjects,
		scope *entity.ObjectScope, requireActive bool, inSettings ...*entity.Setting) error
	LoadRefObjectsByIDs(ctx context.Context, db database.IDB, refObjects **entity.RefObjects,
		scope *entity.ObjectScope, requireActive bool, refIDs *entity.RefObjectIDs) error
	LoadRefObjectsByIDsSkipMissing(ctx context.Context, db database.IDB, refObjects **entity.RefObjects,
		scope *entity.ObjectScope, requireActive bool, refIDs *entity.RefObjectIDs) error

	LoadObjectScope(ctx context.Context, db database.IDB, scopeType base.ObjectScopeType,
		objectID string, requireActive bool) (*entity.ObjectScope, error)
	LoadObjectScopeData(ctx context.Context, db database.IDB, scope *entity.ObjectScope) error

	// Default settings
	InitDefaults(ctx context.Context, db database.IDB) error

	// Events
	OnCreate(ctx context.Context, db database.IDB, event *CreateEvent) error
	OnUpdate(ctx context.Context, db database.IDB, event *UpdateEvent) error
	OnUpdateStatus(ctx context.Context, db database.IDB, event *UpdateEvent) error
	OnDelete(ctx context.Context, db database.IDB, event *DeleteEvent) error
}
