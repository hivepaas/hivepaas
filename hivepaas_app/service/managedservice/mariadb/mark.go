package mariadb

import (
	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/managedservice/managedmodel"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/managedservice/walgkit"
)

// What walg-mark prints, as KEY=value: the GTID the point is at, which a
// restore to the point stops at whatever server's binary logs the archive holds
// by then, and the binary log the point is in with the offset of its end, which
// is the point's position only when it has no GTID - no transaction yet.
// MariaDB keeps no named points: the name is the engine's alone.
const (
	markGTID     = "GTID"
	markPosition = "POSITION"
)

// MarkRestorePoint closes the binary log the point is in and archives it: the
// point is restorable once walg-mark returns.
func (e *Engine) MarkRestorePoint(managed *entity.AppManagedSettings, name string) (
	managedmodel.MarkCommand, error) {
	command, err := walgkit.MarkRestorePoint(managed, name)
	return command, hperrors.Wrap(err)
}

func (e *Engine) MarkResult(stdout string) (string, error) {
	if gtid, err := walgkit.MarkResult(stdout, markGTID, gtidPattern); err == nil {
		return gtid, nil
	}
	position, err := walgkit.MarkResult(stdout, markPosition, positionPattern)
	return position, hperrors.Wrap(err)
}
