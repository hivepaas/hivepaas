package postgres

import (
	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/managedservice/managedmodel"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/managedservice/walgkit"
)

// outPgWalBytes is what Postgres's walg-health says waits on the node to be
// archived: pg_wal's size.
const outPgWalBytes = "PG_WAL_BYTES"

func (e *Engine) Status(req *managedmodel.StatusReq) (*managedmodel.Status, error) {
	status, err := walgkit.Status(profile, req)
	return status, hperrors.Wrap(err)
}
