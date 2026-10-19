package postgres

import (
	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/managedservice/managedmodel"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/managedservice/walgkit"
)

func (e *Engine) NightlyBackupArgs() string {
	return walgkit.NightlyBackupArgs
}

// Retention keeps the retainFull newest full backups, and every one a kept point
// needs (design §4, "Retention and kept points").
func (e *Engine) Retention(req *managedmodel.RetentionReq) (*managedmodel.Retention, error) {
	retention, err := walgkit.Retention(profile, req)
	return retention, hperrors.Wrap(err)
}

func (e *Engine) StaleDeployPoints(managed *entity.AppManagedSettings, outputs map[string]string) []string {
	return walgkit.StaleDeployPoints(profile, managed, outputs)
}
