package getstarteduc

import (
	"context"

	"github.com/hivepaas/hivepaas/hivepaas_app/basedto"
	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/system/getstarteduc/getstarteddto"
)

// Dismiss closes the Get started card, for every admin: what it lists belongs
// to the installation.
func (uc *UC) Dismiss(
	ctx context.Context,
	auth *basedto.Auth,
	_ *getstarteddto.DismissReq,
) (*getstarteddto.DismissResp, error) {
	if err := requireAdmin(auth); err != nil {
		return nil, hperrors.Wrap(err)
	}
	if err := uc.getStartedService.Finish(ctx, uc.db); err != nil {
		return nil, hperrors.Wrap(err)
	}
	return &getstarteddto.DismissResp{Meta: &basedto.Meta{}}, nil
}
