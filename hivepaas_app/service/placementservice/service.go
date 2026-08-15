package placementservice

import (
	"context"

	"github.com/hivepaas/hivepaas/hivepaas_app/infra/database"
)

type Service interface {
	ApplyPlacementSettings(ctx context.Context, db database.IDB, req *ApplyPlacementSettingsReq) (
		*ApplyPlacementSettingsResp, error)
}
