package postgres

import (
	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/managedservice/managedmodel"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/managedservice/walgkit"
)

// Promote is every WAL-G engine's promotion (walgkit.Promote): Postgres's needs
// nothing more.
func (e *Engine) Promote(req *managedmodel.PromoteReq) (*managedmodel.PromoteSpec, error) {
	spec, err := walgkit.Promote(profile, req)
	return spec, hperrors.Wrap(err)
}
