package apptemplateuc

import (
	"context"

	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/infra/database"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/licenseservice"
)

// checkPlanLimit refuses a request that would take the installation past the
// apps its plan allows. It takes the creating transaction, before anything is
// provisioned in it: the check takes the apps' plan-limit lock, which is held
// until the apps are committed.
func (uc *UC) checkPlanLimit(ctx context.Context, db database.Tx, apps []*appToProvision) error {
	return hperrors.Wrap(uc.licenseService.Check(ctx, db,
		licenseservice.Adding(base.LicenseLimitApps, countedApps(apps))))
}

// countedApps is how many of a request's apps count against the plan: the ones
// that belong to no other. The dependencies and components created beside the
// app the request names carry its id as their logical parent, which keeps them
// out of the app lists, and out of the count too. A template makes no function.
func countedApps(apps []*appToProvision) int {
	n := 0
	for _, target := range apps {
		if target.logicalParentID == "" {
			n++
		}
	}
	return n
}
