package syserroruc

import (
	"github.com/hivepaas/hivepaas/hivepaas_app/infra/database"
	"github.com/hivepaas/hivepaas/hivepaas_app/repository"
)

type UC struct {
	db *database.DB

	appErrorRepo repository.SysErrorRepo
}

func New(
	db *database.DB,

	appErrorRepo repository.SysErrorRepo,
) *UC {
	return &UC{
		db: db,

		appErrorRepo: appErrorRepo,
	}
}
