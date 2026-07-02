package taskuc

import (
	"github.com/hivepaas/hivepaas/hivepaas_app/infra/database"
	"github.com/hivepaas/hivepaas/hivepaas_app/repository"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/taskservice"
)

type UC struct {
	db *database.DB

	settingRepo repository.SettingRepo

	taskService taskservice.Service
}

func New(
	db *database.DB,

	settingRepo repository.SettingRepo,

	taskService taskservice.Service,
) *UC {
	return &UC{
		db: db,

		settingRepo: settingRepo,

		taskService: taskService,
	}
}
