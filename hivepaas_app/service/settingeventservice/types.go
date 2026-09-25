package settingeventservice

import "github.com/hivepaas/hivepaas/hivepaas_app/entity"

// Each event's Tasks are what handling it recorded in the caller's transaction,
// for the caller to schedule once that transaction has committed.

type CreateEvent struct {
	Setting *entity.Setting
	Tasks   []*entity.Task
}

type DeleteEvent struct {
	Setting *entity.Setting
	Tasks   []*entity.Task
}

type UpdateEvent struct {
	Setting    *entity.Setting
	OldSetting *entity.Setting
	Tasks      []*entity.Task
}
