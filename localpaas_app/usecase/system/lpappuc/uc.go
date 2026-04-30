package lpappuc

import (
	"github.com/localpaas/localpaas/localpaas_app/infra/database"
	"github.com/localpaas/localpaas/localpaas_app/repository"
	"github.com/localpaas/localpaas/localpaas_app/service/lpappservice"
)

type UC struct {
	db           *database.DB
	lockRepo     repository.LockRepo
	lpAppService lpappservice.Service
}

func New(
	db *database.DB,
	lockRepo repository.LockRepo,
	lpAppService lpappservice.Service,
) *UC {
	return &UC{
		db:           db,
		lockRepo:     lockRepo,
		lpAppService: lpAppService,
	}
}
