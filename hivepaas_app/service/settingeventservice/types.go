package settingeventservice

import "github.com/hivepaas/hivepaas/hivepaas_app/entity"

type CreateEvent struct {
	Setting *entity.Setting
}

type DeleteEvent struct {
	Setting *entity.Setting
}

type UpdateEvent struct {
	Setting    *entity.Setting
	OldSetting *entity.Setting
}
