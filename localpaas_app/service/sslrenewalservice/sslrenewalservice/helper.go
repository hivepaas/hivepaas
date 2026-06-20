package sslrenewalserviceimpl

import (
	"context"

	"github.com/localpaas/localpaas/localpaas_app/apperrors"
	"github.com/localpaas/localpaas/localpaas_app/base"
	"github.com/localpaas/localpaas/localpaas_app/entity"
	"github.com/localpaas/localpaas/localpaas_app/infra/database"
)

func (s *service) sslGetNotification(
	ctx context.Context,
	db database.IDB,
	sslSetting *entity.Setting,
	eventIsSuccess bool,
	data *sslRenewalData,
) (_ *entity.Notification, err error) {
	sslCert := sslSetting.MustAsSSLCert()
	if sslCert.Notification == nil {
		return nil, nil
	}

	data.Mu.Lock()
	defer data.Mu.Unlock()

	var scope *base.ObjectScope
	switch {
	case sslSetting.BelongToApp != nil:
		scope = sslSetting.BelongToApp.GetObjectScope()
	case sslSetting.BelongToProject != nil:
		scope = sslSetting.BelongToProject.GetObjectScope()
	default:
		scope = base.NewObjectScopeGlobal()
	}

	notification, err := s.notificationService.GetNotificationForEvent(ctx, db,
		scope, sslCert.Notification, eventIsSuccess, data.RefObjects)
	if err != nil {
		return nil, apperrors.Wrap(err)
	}
	if notification == nil {
		return nil, nil
	}

	return notification, nil
}
