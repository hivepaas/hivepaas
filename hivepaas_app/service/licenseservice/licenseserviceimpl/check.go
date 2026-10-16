package licenseserviceimpl

import (
	"context"

	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/infra/database"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/license"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/licenseservice"
)

// Check goes through the kinds in base.AllLicenseLimits order, which is also the
// order it takes their locks in: two creations that each add two kinds take the
// locks the same way round, and cannot deadlock.
func (s *service) Check(ctx context.Context, db database.IDB, req *licenseservice.CheckReq) error {
	if req == nil {
		return nil
	}
	adding := false
	for _, limit := range base.AllLicenseLimits {
		if req.Adding[limit] <= 0 {
			continue
		}
		adding = true
		// Required whatever the plan: a caller that passes no transaction is a
		// mistake, and one a paid plan, where nothing is locked, would never
		// show.
		if countedInDB(limit) {
			if _, isTx := db.(database.Tx); !isTx {
				return hperrors.Wrap(hperrors.ErrInternal).
					WithMsgLog("plan check outside a transaction: %s", limit)
			}
		}
	}
	if !adding {
		return nil
	}

	state, err := s.Current(ctx)
	if err != nil {
		return err
	}
	for _, limit := range base.AllLicenseLimits {
		n := req.Adding[limit]
		if n <= 0 {
			continue
		}
		if err = s.checkLimit(ctx, db, state, limit, n); err != nil {
			return err
		}
	}
	return nil
}

// checkLimit refuses adding n of limit when the plan in force would be passed.
// A kind with no limit is passed over before anything is locked or counted.
func (s *service) checkLimit(ctx context.Context, db database.IDB, state *license.State,
	limit base.LicenseLimit, n int) error {
	maxCount, limited := state.Limits.Max(limit)
	if !limited {
		return nil
	}

	var count int
	var err error
	switch limit {
	case base.LicenseLimitSignInProviders:
		// Not a count: the paid kinds that exist keep working on Bee, and the
		// rule is only that none is added.
	case base.LicenseLimitNodes:
		// The swarm's state, which no lock of HivePaaS's serializes, and which
		// anyone with the join token can change without asking it.
		count, err = s.countNodes(ctx)
	case base.LicenseLimitUsers, base.LicenseLimitApps, base.LicenseLimitFunctions, base.LicenseLimitPreviews:
		// Under READ COMMITTED the count starts after the lock is granted, so it
		// sees what the creation that held the lock before committed.
		if err = s.planUsageRepo.LockLimit(ctx, db, limit); err != nil {
			return hperrors.Wrap(err)
		}
		count, err = s.countInDB(ctx, db, limit)
	}
	if err != nil {
		return err
	}

	if count+n > maxCount {
		return planLimitReached(limit, maxCount, state.Plan)
	}
	return nil
}

// countedInDB reports whether limit is counted in the database, under its lock.
func countedInDB(limit base.LicenseLimit) bool {
	switch limit {
	case base.LicenseLimitUsers, base.LicenseLimitApps, base.LicenseLimitFunctions, base.LicenseLimitPreviews:
		return true
	case base.LicenseLimitNodes, base.LicenseLimitSignInProviders:
		return false
	}
	return false
}

// planLimitReached is the refusal. Limit, Max and Plan go on the wire for the
// dashboard's upgrade prompt; the rest is for the message. Max stays an int:
// the message reads differently for a kind the plan leaves out (0).
func planLimitReached(limit base.LicenseLimit, maxCount int, plan base.LicensePlan) error {
	return hperrors.Wrap(hperrors.ErrPlanLimitReached).
		WithParam("Limit", string(limit)).
		WithParam("Max", maxCount).
		WithParam("Plan", string(plan)).
		WithTransParam("LimitName", license.LimitNameMsgID(limit)).
		WithParam("PlanName", plan.DisplayName()).
		ExposeParams("Limit", "Max", "Plan")
}
