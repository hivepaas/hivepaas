package sysstatusuc

import (
	"github.com/hivepaas/hivepaas/hivepaas_app/infra/database"
)

type UC struct {
	db *database.DB
}

func New(db *database.DB) *UC {
	return &UC{db: db}
}
