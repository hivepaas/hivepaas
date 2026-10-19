package postgres

import (
	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/managedservice/managedmodel"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/managedservice/walgkit"
)

// Reconfigure applies what every engine's settings share, then the promotion of
// a member, which is Postgres's.
func (e *Engine) Reconfigure(req *managedmodel.ReconfigureReq) (*managedmodel.ReconfigureResp, error) {
	resp, err := walgkit.Reconfigure(profile, req)
	if err != nil {
		return nil, hperrors.Wrap(err)
	}
	if req.Managed.Database == nil || req.Change == nil {
		return resp, nil
	}
	if err = promoteWriter(req, resp); err != nil {
		return nil, err
	}
	return resp, nil
}
