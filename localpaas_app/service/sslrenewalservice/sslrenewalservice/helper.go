package sslrenewalserviceimpl

import (
	"context"
	"fmt"

	"github.com/go-acme/lego/v5/lego"

	"github.com/localpaas/localpaas/localpaas_app/apperrors"
	"github.com/localpaas/localpaas/localpaas_app/base"
	"github.com/localpaas/localpaas/localpaas_app/config"
	"github.com/localpaas/localpaas/localpaas_app/entity"
	"github.com/localpaas/localpaas/localpaas_app/infra/database"
	"github.com/localpaas/localpaas/services/ssl/acme"
)

func (s *service) sslGetAcmeClient(
	ssl *entity.SSLCert,
	data *sslRenewalData,
) (*acme.Client, error) {
	data.Mu.Lock()
	defer data.Mu.Unlock()

	clientKey := fmt.Sprintf("email:%v:provider:%v", ssl.Email, ssl.Provider.ID)
	if client := data.AcmeClients[clientKey]; client != nil {
		return client, nil
	}

	var provider *entity.SSLProvider
	if ssl.Provider.ID != "" {
		providerSetting := data.RefObjects.RefSettings[ssl.Provider.ID]
		if providerSetting == nil {
			return nil, apperrors.NewNotFound(apperrors.Fmt("SSL provider '%v'", ssl.Provider.ID))
		}
		provider = providerSetting.MustAsSSLProvider()
	}

	http01Provider, err := acme.NewHTTP01Provider(config.Current.DataPathSslAcme().AbsPath())
	if err != nil {
		return nil, apperrors.Wrap(err)
	}

	acmeCfg := acme.ACMEConfig{
		Email:          ssl.Email,
		HTTP01Provider: http01Provider,
	}
	if provider != nil {
		switch ssl.CertType {
		case base.SSLCertTypeLetsEncrypt:
			// Do nothing for now
		case base.SSLCertTypeZeroSSL:
			acmeCfg.CACode = lego.CodeZeroSSL
			acmeCfg.EABKid = provider.ZeroSSL.EABKid
			acmeCfg.EABHmacKey = provider.ZeroSSL.EABHmacKey.MustGetPlain()
		case base.SSLCertTypeGoogleTrust:
			acmeCfg.CACode = lego.CodeGoogleTrust
			acmeCfg.EABKid = provider.GoogleTrust.EABKid
			acmeCfg.EABHmacKey = provider.GoogleTrust.EABHmacKey.MustGetPlain()
		case base.SSLCertTypeSelfSigned, base.SSLCertTypeCustom:
			// Do nothing
		}
	}

	client, err := acme.NewClient(acmeCfg)
	if err != nil {
		return nil, apperrors.Wrap(err)
	}

	// Cache the client
	data.AcmeClients[clientKey] = client

	return client, nil
}

func (s *service) sslGetNotification(
	ctx context.Context,
	db database.IDB,
	sslSetting *entity.Setting,
	eventIsSuccess bool,
	data *sslRenewalData,
) (_ *entity.Notification, err error) {
	ssl := sslSetting.MustAsSSLCert()
	if ssl.Notification == nil {
		return nil, nil
	}

	data.Mu.Lock()
	defer data.Mu.Unlock()

	var scope *base.ObjectScope
	switch {
	case sslSetting.BelongToApp != nil:
		scope = sslSetting.BelongToApp.GetSettingScope()
	case sslSetting.BelongToProject != nil:
		scope = sslSetting.BelongToProject.GetSettingScope()
	default:
		scope = base.NewObjectScopeGlobal()
	}

	notification, err := s.notificationService.GetNotificationForEvent(ctx, db,
		scope, ssl.Notification, eventIsSuccess, data.RefObjects)
	if err != nil {
		return nil, apperrors.Wrap(err)
	}
	if notification == nil {
		return nil, nil
	}

	return notification, nil
}
