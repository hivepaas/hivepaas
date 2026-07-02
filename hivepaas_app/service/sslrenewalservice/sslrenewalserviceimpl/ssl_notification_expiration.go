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
	notification, err := s.sslGetNotification(ctx, db, item.Setting, isSucceeded, data)
	if err != nil {
		return apperrors.New(err)
	}
	if notification == nil {
		return nil
	}

	s.sslBuildExpiringNotificationMsgData(item, data)
	_, err = s.notificationService.NotifyForTaskResult(ctx, db, &notificationservice.TaskResultNotificationReq{
		ActionSucceeded: isSucceeded,
		ScopeProject:    item.Setting.BelongToProject,
		ScopeApp:        item.Setting.BelongToApp,
		RefObjects:      data.RefObjects,
		Notification:    notification,
		TemplateName:    notificationservice.TemplateSSLExpiringNotification,
		TemplateData:    item.ExpiringNotifMsgData,
	})
	if err != nil {
		return apperrors.New(err)
	}
	return nil
}

func (s *service) sslBuildExpiringNotificationMsgData(
	item *sslRenewalDataItem,
	data *sslRenewalData,
) {
	sslCert := item.Setting.MustAsSSLCert()
	project := item.Setting.BelongToProject
	app := item.Setting.BelongToApp
	msgData := &notificationservice.TemplateDataSSLExpiring{
		BaseTemplateData: notificationservice.BaseTemplateData{
			Title: s.notificationService.BuildTitlePrefix(project, app, nil) +
				fmt.Sprintf(" Your SSL expiring in %v", item.ExpiringNotifMsgData.ExpireIn),
		},
		SSLName:   item.Setting.Name,
		SSLType:   string(sslCert.CertType),
		Domain:    sslCert.Domain,
		CreatedAt: item.Setting.CreatedAt.Truncate(time.Second),
		ExpireAt:  sslCert.ExpireAt.Truncate(time.Second),
		ExpireIn:  timeutil.Duration(sslCert.ExpireAt.Sub(timeutil.NowUTC()).Truncate(time.Hour)),
	}
	if project != nil {
		msgData.ProjectName = project.Name
	}
	if app != nil {
		msgData.AppName = app.Name
	}

	switch {
	case app != nil:
		msgData.DashboardLink = config.Current.DashboardAppSchedTaskDetailsURL(app.ProjectID, app.ID,
			data.RenewalJobSetting.ID, data.Task.ID)
	case project != nil:
		msgData.DashboardLink = config.Current.DashboardProjectSchedTaskDetailsURL(project.ID,
			data.RenewalJobSetting.ID, data.Task.ID)
	default:
		msgData.DashboardLink = config.Current.DashboardGlobalSchedTaskDetailsURL(
			data.RenewalJobSetting.ID, data.Task.ID)
	}
	item.ExpiringNotifMsgData = msgData
}
