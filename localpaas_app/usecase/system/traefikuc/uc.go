package traefikuc

import (
	"github.com/localpaas/localpaas/localpaas_app/infra/database"
	"github.com/localpaas/localpaas/localpaas_app/service/traefikservice"
)

type UC struct {
	db             *database.DB
	traefikService traefikservice.Service
}

func New(
	db *database.DB,
	traefikService traefikservice.Service,
) *UC {
	return &UC{
		db:             db,
		traefikService: traefikService,
	}
}
