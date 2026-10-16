package licenseserviceimpl

import (
	"context"

	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/infra/database"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/license"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/licenseservice"
)

// Usage takes no lock: it is for showing, and a creation counts again under its
// own.
func (s *service) Usage(ctx context.Context, db database.IDB) (*licenseservice.Usage, error) {
	used := make(map[base.LicenseLimit]int, len(base.AllLicenseLimits))
	for _, limit := range base.AllLicenseLimits {
		var count int
		var err error
		switch limit {
		case base.LicenseLimitNodes:
			count, err = s.countNodes(ctx)
		case base.LicenseLimitUsers, base.LicenseLimitApps, base.LicenseLimitFunctions, base.LicenseLimitPreviews:
			count, err = s.countInDB(ctx, db, limit)
		case base.LicenseLimitSignInProviders:
			count, err = s.planUsageRepo.CountActiveOAuth(ctx, db, paidSignInKinds())
		}
		if err != nil {
			return nil, hperrors.Wrap(err)
		}
		used[limit] = count
	}
	return &licenseservice.Usage{Used: used}, nil
}

// countNodes counts every node of the swarm, whatever its state: removing a node
// frees its place, draining it does not.
func (s *service) countNodes(ctx context.Context) (int, error) {
	nodes, err := s.dockerManager.NodeList(ctx)
	if err != nil {
		return 0, hperrors.Wrap(err)
	}
	return len(nodes.Items), nil
}

// countInDB counts a kind kept in the database.
func (s *service) countInDB(ctx context.Context, db database.IDB, limit base.LicenseLimit) (int, error) {
	var count int
	var err error
	switch limit {
	case base.LicenseLimitUsers:
		count, err = s.planUsageRepo.CountUsers(ctx, db)
	case base.LicenseLimitApps:
		count, err = s.planUsageRepo.CountApps(ctx, db, false)
	case base.LicenseLimitFunctions:
		count, err = s.planUsageRepo.CountApps(ctx, db, true)
	case base.LicenseLimitPreviews:
		count, err = s.planUsageRepo.CountPreviews(ctx, db)
	case base.LicenseLimitNodes, base.LicenseLimitSignInProviders:
		return 0, hperrors.Wrap(hperrors.ErrInternal).WithMsgLog("%s is not counted in the database", limit)
	}
	if err != nil {
		return 0, hperrors.Wrap(err)
	}
	return count, nil
}

// paidSignInKinds is the sign-in kinds Bee leaves out.
func paidSignInKinds() []base.OAuthKind {
	var kinds []base.OAuthKind
	for _, kind := range base.AllOAuthKinds {
		if license.IsPaidSignIn(kind) {
			kinds = append(kinds, kind)
		}
	}
	return kinds
}
