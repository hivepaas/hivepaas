package sslrenewalserviceimpl

import (
	"context"
	"time"

	"github.com/tiendc/gofn"

	"github.com/hivepaas/hivepaas/hivepaas_app/config"
	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/infra/database"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/timeutil"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/notificationservice"
)

func (s *service) sslNotifyForRenewal(
	ctx context.Context,
	db database.IDB,
	item *sslRenewalDataItem,
	data *sslRenewalData,
) (err error) {
	isSucceeded := item.RenewalError == nil
	notification, err := s.sslGetNotification(ctx, db, item.Scope, item.Setting, isSucceeded, data)
	if err != nil {
		return hperrors.Wrap(err)
	}
	if notification == nil {
		return nil
	}

	s.sslBuildRenewalNotificationMsgData(item, data)
	_, err = s.notificationService.NotifyForTaskResult(ctx, db, &notificationservice.TaskResultNotificationReq{
		ActionSucceeded: isSucceeded,
		Scope:           item.Scope,
		RefObjects:      data.RefObjects,
		Notification:    notification,
		TemplateName:    notificationservice.TemplateSSLRenewalNotification,
		TemplateData:    item.RenewalNotifMsgData,
	})
	if err != nil {
		return hperrors.Wrap(err)
	}
	return nil
}

func (s *service) sslBuildRenewalNotificationMsgData(
	item *sslRenewalDataItem,
	data *sslRenewalData,
) {
	sslCert := item.Setting.MustAsSSLCert()
	timeNow := timeutil.NowUTC()
	isSucceeded := item.RenewalError == nil

	msgData := &notificationservice.TemplateDataSSLRenewal{
		BaseTemplateData: notificationservice.BaseTemplateData{
			Title: s.notificationService.BuildTitlePrefixForScope(item.Scope) +
				gofn.If(isSucceeded, " SSL renewal succeeded", " SSL renewal failed"),
		},
		Succeeded: isSucceeded,
		SSLName:   item.Setting.Name,
		SSLType:   string(sslCert.CertType),
		Domain:    sslCert.Domain,
		CreatedAt: item.Setting.CreatedAt.Truncate(time.Second),
		ExpireAt:  sslCert.ExpireAt.Truncate(time.Second),
	}
	if project := item.Scope.GetProject(); project != nil {
		msgData.ProjectName = project.Name
	}
	if app := item.Scope.GetApp(); app != nil {
		msgData.AppName = app.Name
	}

	if !sslCert.RenewableFrom.IsZero() && sslCert.RenewableFrom.After(timeNow) {
		msgData.NextRenewalIn = timeutil.Duration(sslCert.RenewableFrom.Sub(timeNow).Truncate(time.Hour))
	}

	msgData.DashboardLink = config.Current.DashboardSchedTaskDetailsURL(item.Scope.GetBaseURLPath(),
		data.RenewalJobSetting.ID, data.Task.ID)

	item.RenewalNotifMsgData = msgData
}
