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

	// ReapplyClientIPStrategy regenerates the labels of every app whose routing
	// depends on how a client address is read.
	//
	// Traefik's ip-strategy depth is baked into an app's labels when its routing
	// settings are applied, not read from the proxy settings at request time. So a
	// change to the proxy topology does nothing at all until this runs - it is what
	// makes such a change real, and it has to happen while the change is still on
	// trial rather than afterwards.
	ReapplyClientIPStrategy(ctx context.Context, db database.Tx, req *ReapplyClientIPStrategyReq) (
		*ReapplyClientIPStrategyResp, error)
}
