// Package settingsrevertservice undoes a settings change that nobody confirmed.
//
// It exists as one place because the undo has three callers that must not drift
// apart - the task queue, the timer the app arms for itself, and the reconciler
// that runs at startup - and because what "undo" means depends on which setting
// is on trial. Routing settings go back onto the app's swarm labels; HivePaaS
// service settings go back onto traefik's entrypoint arguments and, when they
// reach that far, onto the service specs themselves.
package settingsrevertservice

import (
	"context"

	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/infra/database"
)

type Service interface {
	// Revert restores the snapshot the change was taken from and re-applies it.
	//
	// It is safe to call more than once and from more than one place at a time:
	// the caller holds the task row, and every reverter refuses to act on a
	// setting whose version has moved on. A revert that does nothing is reported,
	// not treated as a failure.
	Revert(ctx context.Context, db database.Tx, args *entity.TaskSettingsRevertArgs) (*RevertResp, error)
}

type RevertResp struct {
	Reverted bool
	// Reason says why nothing was done, when nothing was done.
	Reason string
}
