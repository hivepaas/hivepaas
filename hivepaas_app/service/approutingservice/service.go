package approutingservice

import (
	"context"

	"github.com/hivepaas/hivepaas/hivepaas_app/infra/database"
)

type Service interface {
	ApplyRoutingSettings(ctx context.Context, db database.IDB, req *ApplyAppRoutingReq) (*ApplyAppRoutingResp, error)

	// RevertSettings restores a routing snapshot and pushes it back onto the app.
	//
	// It is the undo half of confirm-or-revert: the same work ApplyRoutingSettings
	// does, driven from a payload captured before the change rather than from a
	// request. See usecase/system/hpappsettingsuc/routing_settings_probation.go.
	RevertSettings(ctx context.Context, db database.Tx, req *RevertSettingsReq) (*RevertSettingsResp, error)
}
