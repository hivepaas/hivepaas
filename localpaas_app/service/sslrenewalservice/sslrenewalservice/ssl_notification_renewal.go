package sslrenewalserviceimpl

import (
	"context"
	"time"

	"github.com/tiendc/gofn"

	"github.com/localpaas/localpaas/localpaas_app/apperrors"
	"github.com/localpaas/localpaas/localpaas_app/config"
	"github.com/localpaas/localpaas/localpaas_app/infra/database"
	"github.com/localpaas/localpaas/localpaas_app/pkg/timeutil"
	"github.com/localpaas/localpaas/localpaas_app/service/notificationservice"
)

func (s *service) sslNotifyForRenewal(
	ctx context.Context,
	db database.IDB,
	item *sslRenewalDataItem,
	data *sslRenewalData,
) (err error) {
	isSucceeded := item.RenewalError == nil
	notification, err := s.sslGetNotification(ctx, db, item.Setting, isSucceeded, data)
	if err != nil {
		return apperrors.Wrap(err)
	}
	if notification == nil {
		return nil
	}

	s.sslBuildRenewalNotificationMsgData(item, data)
	_, err = s.notificationService.NotifyForTaskResult(ctx, db, &notificationservice.TaskResultNotificationReq{
		ActionSucceeded: isSucceeded,
		ScopeProject:    item.Setting.BelongToProject,
		ScopeApp:        item.Setting.BelongToApp,
		RefObjects:      data.RefObjects,
		Notification:    notification,
		TemplateName:    notificationservice.TemplateSSLRenewalNotification,
		TemplateData:    item.RenewalNotifMsgData,
	})
	if err != nil {
		return apperrors.Wrap(err)
	}
	return nil
}

func (s *service) sslBuildRenewalNotificationMsgData(
	item *sslRenewalDataItem,
	data *sslRenewalData,
) {
	ssl := item.Setting.MustAsSSLCert()
	project := item.Setting.BelongToProject
	app := item.Setting.BelongToApp
	timeNow := timeutil.NowUTC()
	isSucceeded := item.RenewalError == nil

	msgData := &notificationservice.TemplateDataSSLRenewal{
		BaseTemplateData: notificationservice.BaseTemplateData{
			Title: s.notificationService.BuildTitlePrefix(project, app, nil) +
				gofn.If(isSucceeded, " SSL renewal succeeded", " SSL renewal failed"),
		},
		Succeeded: isSucceeded,
		SSLName:   item.Setting.Name,
		SSLType:   string(ssl.CertType),
		Domain:    ssl.Domain,
		CreatedAt: item.Setting.CreatedAt,
		ExpireAt:  ssl.ExpireAt,
	}
	if project != nil {
		msgData.ProjectName = project.Name
	}
	if app != nil {
		msgData.AppName = app.Name
	}

	if !ssl.RenewableFrom.IsZero() && ssl.RenewableFrom.After(timeNow) {
		msgData.NextRenewalIn = timeutil.Duration(ssl.RenewableFrom.Sub(timeNow).Truncate(time.Hour))
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
	item.RenewalNotifMsgData = msgData
}
