package postgres

import (
	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/managedservice/managedmodel"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/managedservice/walgkit"
)

// Reconfigure applies what every engine's settings share, the promotion of a
// member among them.
func (e *Engine) Reconfigure(req *managedmodel.ReconfigureReq) (*managedmodel.ReconfigureResp, error) {
	resp, err := walgkit.Reconfigure(profile, req)
	return resp, hperrors.Wrap(err)
}
