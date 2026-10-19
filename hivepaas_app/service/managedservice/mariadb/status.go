package mariadb

import (
	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/managedservice/managedmodel"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/managedservice/walgkit"
)

// outPendingBytes is what MariaDB's walg-health says waits on the node to be
// archived: the binary logs the server still has, which the archive loop
// purges once pushed.
const outPendingBytes = "PENDING_BYTES"

func (e *Engine) Status(req *managedmodel.StatusReq) (*managedmodel.Status, error) {
	status, err := walgkit.Status(profile, req)
	return status, hperrors.Wrap(err)
}

func (e *Engine) NightlyBackupArgs() string {
	return walgkit.NightlyBackupArgs
}

// Retention keeps the retainFull newest base backups, and every one a kept point
// needs. The binary logs before the oldest kept go with the backups.
func (e *Engine) Retention(req *managedmodel.RetentionReq) (*managedmodel.Retention, error) {
	retention, err := walgkit.Retention(profile, req)
	return retention, hperrors.Wrap(err)
}

func (e *Engine) StaleDeployPoints(managed *entity.AppManagedSettings, outputs map[string]string) []string {
	return walgkit.StaleDeployPoints(profile, managed, outputs)
}

// Reconfigure applies what every engine's settings share. The restore test
// stays in the database's container until the verifier runs MariaDB's.
func (e *Engine) Reconfigure(req *managedmodel.ReconfigureReq) (*managedmodel.ReconfigureResp, error) {
	if req.Change != nil && req.Change.RestoreTest != nil {
		return nil, unsupported("verifier apps")
	}
	resp, err := walgkit.Reconfigure(profile, req)
	return resp, hperrors.Wrap(err)
}
