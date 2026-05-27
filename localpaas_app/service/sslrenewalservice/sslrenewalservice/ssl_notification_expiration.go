package sslrenewalserviceimpl

import (
	"context"
	"fmt"
	"time"

	"github.com/localpaas/localpaas/localpaas_app/apperrors"
	"github.com/localpaas/localpaas/localpaas_app/config"
	"github.com/localpaas/localpaas/localpaas_app/infra/database"
	"github.com/localpaas/localpaas/localpaas_app/pkg/timeutil"
	"github.com/localpaas/localpaas/localpaas_app/service/notificationservice"
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
		return apperrors.Wrap(err)
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
		return apperrors.Wrap(err)
	}
	return nil
}

func (s *service) sslBuildExpiringNotificationMsgData(
	item *sslRenewalDataItem,
	data *sslRenewalData,
) {
	ssl := item.Setting.MustAsSSLCert()
	project := item.Setting.BelongToProject
	app := item.Setting.BelongToApp
	msgData := &notificationservice.TemplateDataSSLExpiring{
		BaseTemplateData: notificationservice.BaseTemplateData{
			Title: s.notificationService.BuildTitlePrefix(project, app, nil) +
				fmt.Sprintf(" Your SSL expiring in %v", item.ExpiringNotifMsgData.ExpireIn),
		},
		SSLName:   item.Setting.Name,
		SSLType:   string(ssl.CertType),
		Domain:    ssl.Domain,
		CreatedAt: item.Setting.CreatedAt,
		ExpireAt:  ssl.ExpireAt,
		ExpireIn:  timeutil.Duration(ssl.ExpireAt.Sub(timeutil.NowUTC()).Truncate(time.Hour)),
	}
	if project != nil {
		msgData.ProjectName = project.Name
	}
	if app != nil {
		msgData.AppName = app.Name
	}

	switch {
	case app != nil:
		msgData.DashboardLink = config.Current.DashboardAppSchedTaskDetailsURL(app.ID, app.ProjectID,
			data.SchedJob.ID, data.Task.ID)
	case project != nil:
		msgData.DashboardLink = config.Current.DashboardProjectSchedTaskDetailsURL(project.ID,
			data.SchedJob.ID, data.Task.ID)
	default:
		msgData.DashboardLink = config.Current.DashboardGlobalSchedTaskDetailsURL(
			data.SchedJob.ID, data.Task.ID)
	}
	item.ExpiringNotifMsgData = msgData
}
