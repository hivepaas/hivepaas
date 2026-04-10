package settingservice

import (
	"context"

	"github.com/localpaas/localpaas/localpaas_app/base"
	"github.com/localpaas/localpaas/localpaas_app/entity"
	"github.com/localpaas/localpaas/localpaas_app/infra/database"
)

type Service interface {
	PersistSettingData(ctx context.Context, db database.IDB, data *PersistingSettingData) error

	LoadReferenceObjects(ctx context.Context, db database.IDB, scope *base.SettingScope, requireActive bool,
		errorIfUnavail bool, inSettings ...*entity.Setting) (*entity.RefObjects, error)
	LoadReferenceObjectsByIDs(ctx context.Context, db database.IDB, scope *base.SettingScope, requireActive bool,
		errorIfUnavail bool, refIDs *entity.RefObjectIDs) (*entity.RefObjects, error)

	// Default settings
	InitDefaults(ctx context.Context, db database.IDB) error

	// Events
	OnCreate(ctx context.Context, db database.IDB, event *CreateEvent) error
	OnUpdate(ctx context.Context, db database.IDB, event *UpdateEvent) error
	OnUpdateStatus(ctx context.Context, db database.IDB, event *UpdateEvent) error
	OnDelete(ctx context.Context, db database.IDB, event *DeleteEvent) error
}
