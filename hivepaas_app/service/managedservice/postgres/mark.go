package postgres

import (
	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/managedservice/managedmodel"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/managedservice/walgkit"
)

const (
	// markLSN is what walg-mark prints, as KEY=value: the point's LSN.
	markLSN = "LSN"
	// pointNameMax is what Postgres keeps of a point's name, in bytes.
	pointNameMax = walgkit.PointNameMax
)

// ValidPointName says whether a restore point may have that name.
func ValidPointName(name string) bool {
	return walgkit.ValidPointName(name)
}

func (e *Engine) MarkRestorePoint(managed *entity.AppManagedSettings, name string) (
	managedmodel.MarkCommand, error) {
	command, err := walgkit.MarkRestorePoint(managed, name)
	return command, hperrors.Wrap(err)
}

func (e *Engine) MarkResult(stdout string) (string, error) {
	lsn, err := walgkit.MarkResult(stdout, markLSN, lsnPattern)
	return lsn, hperrors.Wrap(err)
}
