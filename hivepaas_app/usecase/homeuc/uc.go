package homeuc

import (
	"github.com/hivepaas/hivepaas/hivepaas_app/infra/database"
	"github.com/hivepaas/hivepaas/hivepaas_app/permission"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/attentionservice"
)

// UC serves the home page: what a user lands on, gathered from across the
// system and narrowed to what that user may see.
type UC struct {
	db *database.DB

	permissionManager permission.Manager

	attentionService attentionservice.Service
}

func New(
	db *database.DB,

	permissionManager permission.Manager,

	attentionService attentionservice.Service,
) *UC {
	return &UC{
		db: db,

		permissionManager: permissionManager,

		attentionService: attentionService,
	}
}
