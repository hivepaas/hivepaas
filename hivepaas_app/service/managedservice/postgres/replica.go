package postgres

import (
	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/managedservice/managedmodel"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/managedservice/walgkit"
)

// Replica is a standby the archive seeds, as every WAL-G engine's is
// (walgkit.Replica): Postgres's needs nothing more.
func (e *Engine) Replica(req *managedmodel.ReplicaReq) (*managedmodel.ReplicaSpec, error) {
	spec, err := walgkit.Replica(profile, req)
	return spec, hperrors.Wrap(err)
}
