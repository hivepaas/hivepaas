package hpappuc

import (
	"github.com/hivepaas/hivepaas/hivepaas_app/infra/database"
	"github.com/hivepaas/hivepaas/hivepaas_app/repository"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/hpappservice"
)

type UC struct {
	db *database.DB

	lockRepo repository.LockRepo

	hpAppService hpappservice.Service
}

func New(
	db *database.DB,

	lockRepo repository.LockRepo,

	hpAppService hpappservice.Service,
) *UC {
	return &UC{
		db: db,

		lockRepo: lockRepo,

		hpAppService: hpAppService,
	}
}
