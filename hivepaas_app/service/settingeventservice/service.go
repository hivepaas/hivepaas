package settingeventservice

import (
	"context"

	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/infra/database"
)

type Service interface {
	OnCreate(ctx context.Context, db database.IDB, event *CreateEvent) error
	OnUpdate(ctx context.Context, db database.IDB, event *UpdateEvent) error
	OnUpdateStatus(ctx context.Context, db database.IDB, event *UpdateEvent) error
	OnDelete(ctx context.Context, db database.IDB, event *DeleteEvent) error

	// ScheduleTasks hands the tasks an event recorded to the queue. It is called
	// once the event's transaction has committed.
	ScheduleTasks(ctx context.Context, tasks ...*entity.Task)
}
