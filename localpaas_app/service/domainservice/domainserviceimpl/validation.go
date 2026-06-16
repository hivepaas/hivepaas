package domainserviceimpl

import (
	"context"
	"errors"

	"github.com/localpaas/localpaas/localpaas_app/apperrors"
	"github.com/localpaas/localpaas/localpaas_app/base"
	"github.com/localpaas/localpaas/localpaas_app/infra/database"
	"github.com/localpaas/localpaas/localpaas_app/pkg/bunex"
	"github.com/localpaas/localpaas/localpaas_app/pkg/domainhelper"
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
			return apperrors.New(apperrors.ErrSettingViolation).
				WithParam("Name", apperrors.Fmt("Use of domain '%v'", domain))
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
		return apperrors.NewInUse(apperrors.Fmt("Domain '%v'", conflictDomains[0].DstID))
	}
	return nil
}
