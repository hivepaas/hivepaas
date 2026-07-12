package domainserviceimpl

import (
	"context"
	"errors"

	"github.com/hivepaas/hivepaas/hivepaas_app/apperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/infra/database"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/bunex"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/domainhelper"
)

func (s *service) VerifyProjectDomains(
	ctx context.Context,
	db database.IDB,
	projectID string,
	domains []string,
) error {
	// Load domain settings in project
	domainSetting, err := s.settingRepo.GetSingle(ctx, db, base.NewObjectScopeProject(projectID),
		base.SettingTypeDomainSettings, true)
	if err != nil && !errors.Is(err, apperrors.ErrNotFound) {
		return apperrors.Wrap(err)
	}
	if domainSetting == nil {
		return nil
	}
	domainSettings := domainSetting.MustAsDomainSettings()
	if len(domainSettings.AllowedDomains) == 0 {
		return nil
	}
	for _, domain := range domains {
		if !domainhelper.IsDomainAllowed(domain, domainSettings.AllowedDomains) {
			return apperrors.Wrap(apperrors.ErrDomainUnallowed).WithParam("Domain", domain)
		}
	}

	return nil
}

func (s *service) VerifyDomainsAvailable(
	ctx context.Context,
	db database.IDB,
	domains []string,
	ignoreAppIDs []string,
) error {
	if len(domains) == 0 {
		return nil
	}
	listOpts := []bunex.SelectQueryOption{
		bunex.SelectWhere("res_link.dst_type = ?", base.ResourceTypeDomain),
		bunex.SelectWhereIn("res_link.dst_id IN (?)", domains...),
		bunex.SelectLimit(1),
	}
	if len(ignoreAppIDs) > 0 {
		listOpts = append(listOpts,
			bunex.SelectWhere("res_link.src_type = ?", base.ResourceTypeApp),
			bunex.SelectWhereNotIn("res_link.src_id NOT IN (?)", ignoreAppIDs...),
		)
	}
	conflictDomains, _, err := s.resLinkRepo.List(ctx, db, nil, listOpts...)
	if err != nil {
		return apperrors.Wrap(err)
	}
	if len(conflictDomains) > 0 {
		return apperrors.Wrap(apperrors.ErrDomainInUse).WithParam("Domain", conflictDomains[0].DstID)
	}
	return nil
}
