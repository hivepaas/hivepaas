package appsettingsuc

import (
	"context"
	"errors"
	"time"

	"github.com/tiendc/gofn"

	"github.com/localpaas/localpaas/localpaas_app/apperrors"
	"github.com/localpaas/localpaas/localpaas_app/base"
	"github.com/localpaas/localpaas/localpaas_app/basedto"
	"github.com/localpaas/localpaas/localpaas_app/entity"
	"github.com/localpaas/localpaas/localpaas_app/infra/database"
	"github.com/localpaas/localpaas/localpaas_app/pkg/bunex"
	"github.com/localpaas/localpaas/localpaas_app/pkg/domainhelper"
	"github.com/localpaas/localpaas/localpaas_app/pkg/timeutil"
	"github.com/localpaas/localpaas/localpaas_app/pkg/transaction"
	"github.com/localpaas/localpaas/localpaas_app/pkg/ulid"
	"github.com/localpaas/localpaas/localpaas_app/service/traefikservice"
	"github.com/localpaas/localpaas/localpaas_app/usecase/appsettingsuc/appsettingsdto"
)

func (uc *UC) UpdateAppHttpSettings(
	ctx context.Context,
	auth *basedto.Auth,
	req *appsettingsdto.UpdateAppHttpSettingsReq,
) (*appsettingsdto.UpdateAppHttpSettingsResp, error) {
	err := transaction.Execute(ctx, uc.db, func(db database.Tx) error {
		data := &updateAppHttpSettingsData{}
		err := uc.loadAppHttpSettingsForUpdate(ctx, db, req, data)
		if err != nil {
			return apperrors.Wrap(err)
		}

		persistingData := &persistingAppData{}
		uc.prepareUpdatingAppHttpSettings(ctx, data, persistingData)

		err = uc.persistData(ctx, db, persistingData)
		if err != nil {
			return apperrors.Wrap(err)
		}

		err = uc.applyAppHttpSettings(ctx, data)
		if err != nil {
			return apperrors.Wrap(err)
		}
		return nil
	})
	if err != nil {
		return nil, apperrors.Wrap(err)
	}

	return &appsettingsdto.UpdateAppHttpSettingsResp{}, nil
}

type updateAppHttpSettingsData struct {
	App             *entity.App
	HttpSettings    *entity.Setting
	NewHttpSettings *entity.AppHttpSettings
	RefObjects      *entity.RefObjects

	DomainsToDelete []*entity.ResLink
	DomainsToAdd    []string
}

func (uc *UC) loadAppHttpSettingsForUpdate(
	ctx context.Context,
	db database.Tx,
	req *appsettingsdto.UpdateAppHttpSettingsReq,
	data *updateAppHttpSettingsData,
) error {
	app, err := uc.appService.LoadApp(ctx, db, req.ProjectID, req.AppID, true, true,
		bunex.SelectExcludeColumns(entity.AppDefaultExcludeColumns...),
		bunex.SelectFor("UPDATE OF app"),
		bunex.SelectRelation("Project"),
		bunex.SelectRelation("Settings",
			bunex.SelectWhere("setting.type = ?", base.SettingTypeAppHttp),
		),
		bunex.SelectRelation("DstResLinks"),
	)
	if err != nil {
		return apperrors.Wrap(err)
	}
	data.App = app
	data.HttpSettings, _ = gofn.First(app.Settings)

	if data.HttpSettings != nil && data.HttpSettings.UpdateVer != req.UpdateVer {
		return apperrors.Wrap(apperrors.ErrUpdateVerMismatched)
	}

	newHttpSettings := req.ToEntity()
	data.NewHttpSettings = newHttpSettings

	// Make sure all reference settings used in this settings exist actively
	data.RefObjects, err = uc.settingService.LoadReferenceObjectsByIDs(ctx, db, app.GetSettingScope(),
		true, true, newHttpSettings.GetRefObjectIDs())
	if err != nil {
		return apperrors.Wrap(err)
	}

	// Active domains of the app need to validate
	activeDomains := newHttpSettings.GetActiveDomainNames()

	// Load domain settings in project
	domainSttg, err := uc.settingRepo.GetSingle(ctx, db, app.GetSettingScope(),
		base.SettingTypeDomainSettings, true)
	if err != nil && !errors.Is(err, apperrors.ErrNotFound) {
		return apperrors.Wrap(err)
	}
	for domainSttg != nil {
		domainSettings := domainSttg.MustAsDomainSettings()
		if len(domainSettings.AllowedDomains) == 0 {
			break
		}
		for _, domain := range activeDomains {
			if !domainhelper.IsDomainAllowed(domain, domainSettings.AllowedDomains) {
				return apperrors.New(apperrors.ErrSettingViolated).
					WithParam("Name", apperrors.Fmt("Use of domain '%v'", domain))
			}
		}
		break //nolint:staticcheck
	}

	// Make sure all domains used by the app are not hold by any other app
	if len(activeDomains) > 0 {
		conflictDomains, _, err := uc.resLinkRepo.List(ctx, db, nil,
			bunex.SelectWhere("res_link.src_type = ?", base.ResourceTypeApp),
			bunex.SelectWhere("res_link.src_id != ?", app.ID),
			bunex.SelectWhere("res_link.dst_type = ?", base.ResourceTypeDomain),
			bunex.SelectWhereIn("res_link.dst_id IN (?)", activeDomains...),
			bunex.SelectLimit(1),
		)
		if err != nil {
			return apperrors.Wrap(err)
		}
		if len(conflictDomains) > 0 {
			return apperrors.NewInUse(apperrors.Fmt("Domain '%v'", conflictDomains[0].DstID))
		}
	}

	// Calculate domain links update
	mapCurrentDomainLinks := make(map[string]*entity.ResLink, len(app.DstResLinks))
	for _, domainLink := range app.DstResLinks {
		if domainLink.DstType != base.ResourceTypeDomain {
			continue
		}
		mapCurrentDomainLinks[domainLink.DstID] = domainLink
	}
	for _, domain := range newHttpSettings.GetActiveDomainNames() {
		domainLink := mapCurrentDomainLinks[domain]
		if domainLink == nil {
			data.DomainsToAdd = append(data.DomainsToAdd, domain)
		} else {
			delete(mapCurrentDomainLinks, domain)
		}
	}
	for _, domainLink := range mapCurrentDomainLinks {
		data.DomainsToDelete = append(data.DomainsToDelete, domainLink)
	}

	return nil
}

func (uc *UC) prepareUpdatingAppHttpSettings(
	_ context.Context,
	data *updateAppHttpSettingsData,
	persistingData *persistingAppData,
) {
	app := data.App
	setting := data.HttpSettings
	timeNow := timeutil.NowUTC()

	if setting == nil {
		setting = &entity.Setting{
			ID:        gofn.Must(ulid.NewStringULID()),
			Scope:     base.ObjectScopeApp,
			ObjectID:  app.ID,
			Type:      base.SettingTypeAppHttp,
			CreatedAt: timeNow,
			Version:   entity.CurrentAppHttpSettingsVersion,
		}
		data.HttpSettings = setting
	}
	setting.UpdateVer++
	setting.UpdatedAt = timeNow
	setting.Status = base.SettingStatusActive
	setting.ExpireAt = time.Time{}
	setting.MustSetData(data.NewHttpSettings)
	persistingData.UpsertingSettings = append(persistingData.UpsertingSettings, setting)

	// Domain links
	for _, domainLink := range data.DomainsToDelete {
		domainLink.DeletedAt = timeNow
		persistingData.UpsertingResLinks = append(persistingData.UpsertingResLinks, domainLink)
	}
	for _, domain := range data.DomainsToAdd {
		persistingData.UpsertingResLinks = append(persistingData.UpsertingResLinks, &entity.ResLink{
			SrcType: base.ResourceTypeApp,
			SrcID:   app.ID,
			DstType: base.ResourceTypeDomain,
			DstID:   domain,
		})
	}
}

func (uc *UC) applyAppHttpSettings(
	ctx context.Context,
	data *updateAppHttpSettingsData,
) error {
	appHttpSettings, err := data.HttpSettings.AsAppHttpSettings()
	if err != nil {
		return apperrors.Wrap(err)
	}

	sslSettings := map[string]*entity.Setting{}
	for _, sslID := range appHttpSettings.GetSSLCertIDs() {
		if s := data.RefObjects.RefSettings[sslID]; s != nil {
			sslSettings[s.ID] = s
		}
	}
	err = uc.sslService.WriteCertFiles(false, gofn.MapValues(sslSettings)...)
	if err != nil {
		return apperrors.Wrap(err)
	}

	inspect, err := uc.dockerManager.ServiceInspect(ctx, data.App.ServiceID)
	if err != nil {
		return apperrors.Wrap(err)
	}
	service := &inspect.Service

	err = uc.traefikService.ApplyAppConfig(ctx, data.App, service, &traefikservice.AppConfigData{
		HttpSettings: appHttpSettings,
		RefObjects:   data.RefObjects,
	})
	if err != nil {
		return apperrors.Wrap(err)
	}

	err = uc.networkService.UpdateAppGlobalRoutingNetwork(ctx, data.App, service, data.HttpSettings)
	if err != nil {
		return apperrors.Wrap(err)
	}

	_, err = uc.dockerManager.ServiceUpdate(ctx, service.ID, &service.Version, &service.Spec)
	if err != nil {
		return apperrors.Wrap(err)
	}

	return nil
}
