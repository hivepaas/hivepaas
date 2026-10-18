package licenseserviceimpl

import (
	"context"

	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/license"
)

func (s *service) CheckFeatures(ctx context.Context, features []string) error {
	if len(features) == 0 {
		return nil
	}
	state, err := s.Current(ctx)
	if err != nil {
		return err
	}
	for _, feature := range features {
		if !state.Grants(feature) {
			return planFeatureRequired(feature, state.Plan)
		}
	}
	return nil
}

// planFeatureRequired is the refusal. Feature and Plan go on the wire for the
// dashboard's upgrade prompt, as a limit's do.
func planFeatureRequired(feature string, plan base.LicensePlan) error {
	err := hperrors.Wrap(hperrors.ErrPlanFeatureRequired).
		WithParam("Feature", feature).
		WithParam("Plan", string(plan)).
		WithParam("PlanName", plan.DisplayName()).
		ExposeParams("Feature", "Plan")
	if msgID := license.FeatureNameMsgID(base.LicenseFeature(feature)); msgID != "" {
		return err.WithTransParam("FeatureName", msgID)
	}
	return err.WithParam("FeatureName", feature)
}
