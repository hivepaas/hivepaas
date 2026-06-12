package sslrenewalserviceimpl

import (
	"context"
	"fmt"

	"github.com/tiendc/gofn"

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

	email := ssl.Email
	keyType := gofn.Coalesce(ssl.KeyType, base.SSLKeyTypeDefault)
	clientKey := fmt.Sprintf("email:%v:keyType:%v:provider:%v", email, keyType, ssl.Provider.ID)

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

	acmeCfg := acme.ACMEConfig{
		Email:         email,
		KeyType:       keyType,
		HTTP01WebRoot: config.Current.DataPathSslLetsEncrypt().AbsPath(),
	}

	if provider != nil {
		switch ssl.CertType {
		case base.SSLCertTypeLetsEncrypt:
			// Do nothing for now
		case base.SSLCertTypeZeroSSL:
			acmeCfg.CADirURL = base.ZeroSSLACMEURL
			acmeCfg.EABKid = provider.ZeroSSL.EABKid
			acmeCfg.EABHmacKey = provider.ZeroSSL.EABHmacKey.MustGetPlain()
		case base.SSLCertTypeGoogleTS:
			acmeCfg.CADirURL = base.GoogleTSACMEURL
			acmeCfg.EABKid = provider.GoogleTS.EABKid
			acmeCfg.EABHmacKey = provider.GoogleTS.EABHmacKey.MustGetPlain()
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
