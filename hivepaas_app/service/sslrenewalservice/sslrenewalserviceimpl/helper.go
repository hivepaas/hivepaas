package sslrenewalserviceimpl

import (
	"context"

	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/infra/database"
)

func (s *service) sslGetNotification(
	ctx context.Context,
	db database.IDB,
	scope *entity.ObjectScope,
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

	notification, err := s.notificationService.GetNotificationForEvent(ctx, db,
		scope, sslCert.Notification, eventIsSuccess, data.RefObjects)
	if err != nil {
		return nil, hperrors.Wrap(err)
	}
	if notification == nil {
		return nil, nil
	}

	return notification, nil
}
