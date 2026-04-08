package appuc

import (
	"context"

	"github.com/go-acme/lego/v4/certificate"
	"github.com/tiendc/gofn"

	"github.com/localpaas/localpaas/localpaas_app/apperrors"
	"github.com/localpaas/localpaas/localpaas_app/base"
	"github.com/localpaas/localpaas/localpaas_app/basedto"
	"github.com/localpaas/localpaas/localpaas_app/config"
	"github.com/localpaas/localpaas/localpaas_app/entity"
	"github.com/localpaas/localpaas/localpaas_app/infra/database"
	"github.com/localpaas/localpaas/localpaas_app/pkg/bunex"
	"github.com/localpaas/localpaas/localpaas_app/pkg/timeutil"
	"github.com/localpaas/localpaas/localpaas_app/pkg/transaction"
	"github.com/localpaas/localpaas/localpaas_app/pkg/ulid"
	"github.com/localpaas/localpaas/localpaas_app/service/traefikservice"
	"github.com/localpaas/localpaas/localpaas_app/usecase/appuc/appdto"
	"github.com/localpaas/localpaas/services/ssl/letsencrypt"
)

func (uc *UC) ObtainDomainSSL(
	ctx context.Context,
	auth *basedto.Auth,
	req *appdto.ObtainDomainSSLReq,
) (*appdto.ObtainDomainSSLResp, error) {
	email := gofn.Coalesce(req.Email, config.Current.SSL.LeUserEmail)
	leClient, err := letsencrypt.NewClient(email, req.KeyType, config.Current.DataPathSslLetsEncrypt())
	if err != nil {
		return nil, apperrors.Wrap(err)
	}

	certificates, renewalInfo, err := leClient.ObtainCertificateWithDetails(ctx, []string{req.Domain})
	if err != nil {
		return nil, apperrors.Wrap(err)
	}
	appData := &obtainSSLData{
		ObtainedCerts: certificates,
		RenewalInfo:   renewalInfo,
		Email:         email,
	}

	err = transaction.Execute(ctx, uc.db, func(db database.Tx) error {
		err = uc.loadAppDataForObtainDomainSSL(ctx, db, req, appData)
		if err != nil {
			return apperrors.Wrap(err)
		}

		persistingData := &persistingAppData{}
		uc.preparePersistingDomainSSLData(req, appData, persistingData)

		err = uc.appService.PersistAppData(ctx, db, &persistingData.PersistingAppData)
		if err != nil {
			return apperrors.Wrap(err)
		}

		err = uc.applyDomainSSL(ctx, db, appData)
		if err != nil {
			return apperrors.Wrap(err)
		}

		return nil
	})
	if err != nil {
		return nil, apperrors.Wrap(err)
	}

	return &appdto.ObtainDomainSSLResp{
		Data: &basedto.ObjectIDResp{ID: appData.SSLCert.ID},
	}, nil
}

type obtainSSLData struct {
	App           *entity.App
	HttpSettings  *entity.Setting
	SSLCert       *entity.Setting
	ObtainedCerts *certificate.Resource
	RenewalInfo   *certificate.RenewalInfoResponse
	Email         string
}

func (uc *UC) loadAppDataForObtainDomainSSL(
	ctx context.Context,
	db database.IDB,
	req *appdto.ObtainDomainSSLReq,
	data *obtainSSLData,
) error {
	app, err := uc.appService.LoadApp(ctx, db, req.ProjectID, req.AppID, true, true,
		bunex.SelectExcludeColumns(entity.AppDefaultExcludeColumns...),
		bunex.SelectFor("UPDATE OF app"),
		bunex.SelectRelation("Project"),
		bunex.SelectRelation("Settings",
			bunex.SelectWhere("setting.type = ?", base.SettingTypeAppHttp),
		),
	)
	if err != nil {
		return apperrors.Wrap(err)
	}
	data.App = app

	httpSettings := app.GetSettingByType(base.SettingTypeAppHttp)
	if httpSettings == nil {
		return apperrors.NewNotFound("AppHttpSetting")
	}

	domainSettings := httpSettings.MustAsAppHttpSettings().GetDomain(req.Domain)
	if domainSettings != nil && domainSettings.SSLCert.ID != "" {
		return apperrors.New(apperrors.ErrAlreadyExist).
			WithMsgLog("ssl for domain '%s' already exists", req.Domain)
	}

	data.HttpSettings = httpSettings

	return nil
}

func (uc *UC) preparePersistingDomainSSLData(
	req *appdto.ObtainDomainSSLReq,
	data *obtainSSLData,
	persistingData *persistingAppData,
) {
	timeNow := timeutil.NowUTC()
	dbSSL := &entity.Setting{
		ID:        gofn.Must(ulid.NewStringULID()),
		Scope:     base.SettingScopeProject,
		ObjectID:  data.App.ProjectID,
		Type:      base.SettingTypeSSLCert,
		Status:    base.SettingStatusActive,
		Name:      req.Domain,
		Kind:      string(base.SSLCertTypeLetsEncrypt),
		CreatedAt: timeNow,
		UpdatedAt: timeNow,
	}
	data.SSLCert = dbSSL

	ssl := &entity.SSLCert{
		Domain:      req.Domain,
		Certificate: string(data.ObtainedCerts.Certificate),
		PrivateKey:  entity.NewEncryptedField(string(data.ObtainedCerts.PrivateKey)),
		KeyType:     req.KeyType,
		CertType:    base.SSLCertTypeLetsEncrypt,
		Email:       data.Email,
	}
	if data.RenewalInfo != nil {
		ssl.RenewableFrom = data.RenewalInfo.SuggestedWindow.Start.UTC()
		if ssl.CertType == base.SSLCertTypeLetsEncrypt && !ssl.RenewableFrom.IsZero() {
			ssl.AutoRenew = true
			// TODO: need a better method to have expiration date of SSLs from Let's encrypt.
			ssl.ExpireAt = ssl.RenewableFrom.Add(base.LetsEncryptExpirationFromFirstRenewableDate)
		}
	}

	dbSSL.MustSetData(ssl)
	persistingData.UpsertingSettings = append(persistingData.UpsertingSettings, dbSSL)

	httpSettings := data.HttpSettings.MustAsAppHttpSettings()
	domainSettings := httpSettings.GetDomain(req.Domain)
	if domainSettings == nil {
		domainSettings = &entity.AppDomain{
			Enabled: true,
			Domain:  req.Domain,
		}
		httpSettings.Domains = append(httpSettings.Domains, domainSettings)
	}
	domainSettings.SSLCert.ID = dbSSL.ID

	// Enables the HTTP settings
	httpSettings.ExposePublicly = true
	data.HttpSettings.MustSetData(httpSettings)

	persistingData.UpsertingSettings = append(persistingData.UpsertingSettings, data.HttpSettings)
}

func (uc *UC) applyDomainSSL(
	ctx context.Context,
	db database.IDB,
	data *obtainSSLData,
) error {
	appHttpSettings, err := data.HttpSettings.AsAppHttpSettings()
	if err != nil {
		return apperrors.Wrap(err)
	}

	refObjects, err := uc.settingService.LoadReferenceObjects(ctx, db, data.App.GetSettingScope(),
		true, true, data.HttpSettings)
	if err != nil {
		return apperrors.Wrap(err)
	}

	sslSettings := map[string]*entity.Setting{}
	for _, sslID := range appHttpSettings.GetSSLCertIDs() {
		if s := refObjects.RefSettings[sslID]; s != nil {
			sslSettings[s.ID] = s
		}
	}
	err = uc.sslService.WriteCertFiles(false, gofn.MapValues(sslSettings)...)
	if err != nil {
		return apperrors.Wrap(err)
	}

	service, err := uc.dockerManager.ServiceInspect(ctx, data.App.ServiceID)
	if err != nil {
		return apperrors.Wrap(err)
	}

	err = uc.traefikService.ApplyAppConfig(ctx, data.App, service, &traefikservice.AppConfigData{
		HttpSettings: appHttpSettings,
		RefObjects:   refObjects,
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
