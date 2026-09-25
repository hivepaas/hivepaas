package appprovisionserviceimpl

import (
	"context"
	"errors"

	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/infra/database"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/appprovisionservice"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/clusterservice"
)

func (s *service) ProvisionApps(
	ctx context.Context,
	db database.IDB,
	req *appprovisionservice.ProvisionAppsReq,
) (*appprovisionservice.ProvisionAppsResp, error) {
	resp := &appprovisionservice.ProvisionAppsResp{
		Apps: make([]*appprovisionservice.ProvisionAppResp, 0, len(req.Apps)),
	}
	resp.Cleanup = func(cleanupCtx context.Context) error {
		return s.removeCreated(cleanupCtx, resp.Apps)
	}

	for _, appReq := range req.Apps {
		one, err := s.ProvisionApp(ctx, db, appReq)
		if one != nil {
			resp.Apps = append(resp.Apps, one)
		}
		if err != nil {
			return resp, hperrors.Wrap(err).WithExtraDetail("while creating %s", appReq.Name)
		}
	}
	return resp, nil
}

// removeCreated undoes in docker what provisioning made, newest app first. Their
// records are gone, so anything left here is running or stored with nothing
// pointing at it, and an error names what has to be removed by hand.
func (s *service) removeCreated(ctx context.Context, apps []*appprovisionservice.ProvisionAppResp) error {
	var errs []error
	for i := len(apps) - 1; i >= 0; i-- {
		created := apps[i].Created
		if created == nil {
			continue
		}
		name := "an app"
		if apps[i].App != nil {
			name = apps[i].App.Name
		}
		if created.ServiceID != "" {
			err := s.clusterService.ServiceRemove(ctx, created.ServiceID, clusterservice.ItemRemovalRetryMax, 0)
			if err != nil {
				errs = append(errs, hperrors.Wrap(err).WithExtraDetail(
					"%s was created and could not be removed: remove service %s by hand", name, created.ServiceID))
			}
		}
		if err := s.removeCreatedFiles(ctx, created); err != nil {
			errs = append(errs, hperrors.Wrap(err).WithExtraDetail(
				"%s was created and what it kept in docker could not be removed", name))
		}
	}
	return errors.Join(errs...)
}

// removeCreatedFiles removes the objects of one app's setting mounts, made
// with its service.
func (s *service) removeCreatedFiles(ctx context.Context, created *appprovisionservice.CreatedInDocker) error {
	if created.AppID == "" {
		return nil
	}
	return hperrors.Wrap(s.settingMountService.RemoveApp(ctx, created.AppID))
}
