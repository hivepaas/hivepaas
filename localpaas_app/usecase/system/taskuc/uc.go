package taskuc

import (
	"github.com/localpaas/localpaas/localpaas_app/infra/database"
	"github.com/localpaas/localpaas/localpaas_app/repository"
	"github.com/localpaas/localpaas/localpaas_app/service/taskservice"
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
