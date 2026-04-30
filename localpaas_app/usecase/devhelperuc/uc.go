package devhelperuc

import (
	"github.com/localpaas/localpaas/localpaas_app/infra/database"
	"github.com/localpaas/localpaas/localpaas_app/repository"
)

type UC struct {
	db       *database.DB
	taskRepo repository.TaskRepo
}

func New(
	db *database.DB,
	taskRepo repository.TaskRepo,
) *UC {
	return &UC{
		db:       db,
		taskRepo: taskRepo,
	}
}
