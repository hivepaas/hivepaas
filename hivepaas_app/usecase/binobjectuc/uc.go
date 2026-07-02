package binobjectuc

import (
	"github.com/hivepaas/hivepaas/hivepaas_app/infra/database"
	"github.com/hivepaas/hivepaas/hivepaas_app/repository"
)

type UC struct {
	db *database.DB

	binObjectRepo repository.BinObjectRepo
}

func New(
	db *database.DB,

	binObjectRepo repository.BinObjectRepo,
) *UC {
	return &UC{
		db: db,

		binObjectRepo: binObjectRepo,
	}
}
