package appuc

import (
	"context"

	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/infra/database"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/licenseservice"
)

// checkPlanLimit refuses creating one more app or function than the plan in
// force allows. It takes the creating transaction, first thing in it: the check
// takes the kind's plan-limit lock, which is held until the creation commits, so
// two creations at once cannot both count one place left.
func (uc *UC) checkPlanLimit(ctx context.Context, db database.Tx, limit base.LicenseLimit) error {
	return hperrors.Wrap(uc.licenseService.Check(ctx, db, licenseservice.Adding(limit, 1)))
}
