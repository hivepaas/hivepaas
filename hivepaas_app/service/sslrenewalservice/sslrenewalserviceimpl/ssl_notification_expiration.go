package sslrenewalserviceimpl

import (
	"context"
	"fmt"
	"time"

	"github.com/hivepaas/hivepaas/hivepaas_app/apperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/config"
	"github.com/hivepaas/hivepaas/hivepaas_app/infra/database"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/timeutil"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/notificationservice"
)

func (s *service) sslNotifyForExpiration(
	ctx context.Context,
	db database.IDB,
	item *sslRenewalDataItem,
	data *sslRenewalData,
) (err error) {
	isSucceeded := false
	notification, err := s.sslGetNotification(ctx, db, item.Scope, item.Setting, isSucceeded, data)
	if err != nil {
		return apperrors.Wrap(err)
	}
	if notification == nil {
		return nil
	}

	s.sslBuildExpiringNotificationMsgData(item, data)
	_, err = s.notificationService.NotifyForTaskResult(ctx, db, &notificationservice.TaskResultNotificationReq{
		ActionSucceeded: isSucceeded,
		Scope:           item.Scope,
		RefObjects:      data.RefObjects,
		Notification:    notification,
		TemplateName:    notificationservice.TemplateSSLExpiringNotification,
		TemplateData:    item.ExpiringNotifMsgData,
	})
	if err != nil {
		return apperrors.Wrap(err)
	}
	return nil
}

func (s *service) sslBuildExpiringNotificationMsgData(
	item *sslRenewalDataItem,
	data *sslRenewalData,
) {
	sslCert := item.Setting.MustAsSSLCert()
	msgData := &notificationservice.TemplateDataSSLExpiring{
		BaseTemplateData: notificationservice.BaseTemplateData{
			Title: s.notificationService.BuildTitlePrefixForScope(item.Scope) +
				fmt.Sprintf(" Your SSL expiring in %v", item.ExpiringNotifMsgData.ExpireIn),
		},
		SSLName:   item.Setting.Name,
		SSLType:   string(sslCert.CertType),
		Domain:    sslCert.Domain,
		CreatedAt: item.Setting.CreatedAt.Truncate(time.Second),
		ExpireAt:  sslCert.ExpireAt.Truncate(time.Second),
		ExpireIn:  timeutil.Duration(sslCert.ExpireAt.Sub(timeutil.NowUTC()).Truncate(time.Hour)),
	}
	if project := item.Scope.GetProject(); project != nil {
		msgData.ProjectName = project.Name
	}
	if app := item.Scope.GetApp(); app != nil {
		msgData.AppName = app.Name
	}

	msgData.DashboardLink = config.Current.DashboardSchedTaskDetailsURL(item.Scope.GetBaseURLPath(),
		data.RenewalJobSetting.ID, data.Task.ID)

	item.ExpiringNotifMsgData = msgData
}
