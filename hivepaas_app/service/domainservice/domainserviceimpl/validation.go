package domainserviceimpl

import (
	"context"
	"errors"

	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
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
	domainSetting, err := s.settingRepo.GetSingle(ctx, db, entity.NewObjectScopeProject(projectID),
		base.SettingTypeDomainSettings, true)
	if err != nil && !errors.Is(err, hperrors.ErrNotFound) {
		return hperrors.Wrap(err)
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
			return hperrors.Wrap(hperrors.ErrDomainUnallowed).WithParam("Domain", domain)
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
	// A domain is held by the setting that records it, never by the app itself,
	// so both of these ask about the setting behind the link. One that is gone
	// holds nothing: a link outliving its setting would otherwise make a domain
	// unusable by anything, with no way to see why.
	listOpts := []bunex.SelectQueryOption{
		bunex.SelectWhere("res_link.dst_type = ?", base.ResourceTypeDomain),
		bunex.SelectWhereIn("res_link.dst_id IN (?)", domains...),
		bunex.SelectWhere("EXISTS (SELECT 1 FROM settings" +
			" WHERE settings.id = res_link.src_id AND settings.deleted_at IS NULL)"),
		bunex.SelectLimit(1),
	}
	if len(ignoreAppIDs) > 0 {
		listOpts = append(listOpts,
			bunex.SelectWhere("NOT EXISTS (SELECT 1 FROM settings"+
				" WHERE settings.id = res_link.src_id AND settings.scope = ?"+
				" AND settings.object_id IN (?))",
				base.ObjectScopeApp, bunex.List(ignoreAppIDs)),
		)
	}
	conflictDomains, _, err := s.resLinkRepo.List(ctx, db, nil, listOpts...)
	if err != nil {
		return hperrors.Wrap(err)
	}
	if len(conflictDomains) > 0 {
		return hperrors.Wrap(hperrors.ErrDomainInUse).WithParam("Domain", conflictDomains[0].DstID)
	}
	return nil
}
