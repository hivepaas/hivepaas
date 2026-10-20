package mariadb

import (
	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/managedservice/managedmodel"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/managedservice/walgkit"
)

// markPosition is what walg-mark prints, as KEY=value: the binary log the point
// is in and the offset of its end, which a restore to the point replays to.
// MariaDB keeps no named points: the name is the engine's alone.
const markPosition = "POSITION"

// MarkRestorePoint closes the binary log the point is in and archives it: the
// point is restorable once walg-mark returns.
func (e *Engine) MarkRestorePoint(managed *entity.AppManagedSettings, name string) (
	managedmodel.MarkCommand, error) {
	command, err := walgkit.MarkRestorePoint(managed, name)
	return command, hperrors.Wrap(err)
}

func (e *Engine) MarkResult(stdout string) (string, error) {
	position, err := walgkit.MarkResult(stdout, markPosition, positionPattern)
	return position, hperrors.Wrap(err)
}
