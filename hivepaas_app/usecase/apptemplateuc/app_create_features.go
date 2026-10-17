package apptemplateuc

import (
	"context"
	"slices"

	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
)

// checkFeatures refuses a request needing a license feature the plan in force
// does not include, before anything is created for it. What every app of the
// request needs is needed: a dependency is created by the same person, in the
// same request. A request needing none asks the license nothing.
func (uc *UC) checkFeatures(ctx context.Context, apps []*appToProvision) error {
	var features []string
	for _, target := range apps {
		if target.rendered == nil || target.rendered.Template == nil {
			continue
		}
		features = append(features, target.rendered.Template.Metadata.Requires.Features...)
	}
	if len(features) == 0 {
		return nil
	}
	slices.Sort(features)
	return hperrors.Wrap(uc.licenseService.CheckFeatures(ctx, slices.Compact(features)))
}
