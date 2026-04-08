package syserroruc

import (
	"github.com/localpaas/localpaas/localpaas_app/infra/database"
	"github.com/localpaas/localpaas/localpaas_app/repository"
)

type UC struct {
	db           *database.DB
	appErrorRepo repository.SysErrorRepo
}

func New(
	db *database.DB,
	appErrorRepo repository.SysErrorRepo,
) *UC {
	return &UC{
		db:           db,
		appErrorRepo: appErrorRepo,
	}
}
