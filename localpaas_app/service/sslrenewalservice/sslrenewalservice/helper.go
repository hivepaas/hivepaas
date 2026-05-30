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
	"github.com/localpaas/localpaas/services/ssl/letsencrypt"
)

func (s *service) sslGetLeClient(
	ssl *entity.SSLCert,
	data *sslRenewalData,
) (*letsencrypt.Client, error) {
	data.Mu.Lock()
	defer data.Mu.Unlock()

	email := ssl.Email
	keyType := gofn.Coalesce(ssl.KeyType, base.SSLKeyTypeDefault)
	mapKey := fmt.Sprintf("email:%v:keysize:%v", email, keyType)

	if client := data.LeClients[mapKey]; client != nil {
		return client, nil
	}

	client, err := letsencrypt.NewClient(email, keyType, config.Current.DataPathSslLetsEncrypt())
	if err != nil {
		return nil, apperrors.Wrap(err)
	}
	data.LeClients[mapKey] = client

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
