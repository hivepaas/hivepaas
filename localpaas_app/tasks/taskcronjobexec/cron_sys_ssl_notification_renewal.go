package taskcronjobexec

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

func (e *Executor) sslNotifyOfRenewal(
	ctx context.Context,
	db database.IDB,
	item *sslRenewalTaskItem,
	data *sslRenewalTaskData,
) (err error) {
	isSucceeded := item.RenewalError == nil
	notification, err := e.sslGetNotification(ctx, db, item.Setting, isSucceeded, data)
	if err != nil {
		return apperrors.Wrap(err)
	}
	if notification == nil {
		return nil
	}

	e.sslBuildRenewalNotificationMsgData(item, data)
	_, err = e.notificationService.NotifyForTaskResult(ctx, db, &notificationservice.TaskResultNotificationReq{
		ActionSucceeded: isSucceeded,
		ScopeProject:    data.Project,
		ScopeApp:        data.App,
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

func (e *Executor) sslBuildRenewalNotificationMsgData(
	item *sslRenewalTaskItem,
	data *sslRenewalTaskData,
) {
	ssl := item.Setting.MustAsSSLCert()
	timeNow := timeutil.NowUTC()
	isSucceeded := item.RenewalError == nil
	msgData := &notificationservice.TemplateDataSSLRenewal{
		BaseTemplateData: notificationservice.BaseTemplateData{
			Title: e.notificationService.BuildTitlePrefix(data.Project, data.App, nil) +
				gofn.If(isSucceeded, " SSL renewal succeeded", " SSL renewal failed"),
		},
		Succeeded: isSucceeded,
		SSLName:   item.Setting.Name,
		SSLType:   string(ssl.CertType),
		Domain:    ssl.Domain,
		CreatedAt: item.Setting.CreatedAt,
		ExpireAt:  ssl.ExpireAt,
	}
	project := item.Setting.BelongToProject
	app := item.Setting.BelongToApp
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
		msgData.DashboardLink = config.Current.DashboardAppCronTaskDetailsURL(app.ID, app.ProjectID,
			data.CronJobSetting.ID, data.Task.ID)
	case project != nil:
		msgData.DashboardLink = config.Current.DashboardProjectCronTaskDetailsURL(project.ID,
			data.CronJobSetting.ID, data.Task.ID)
	default:
		msgData.DashboardLink = config.Current.DashboardGlobalCronTaskDetailsURL(
			data.CronJobSetting.ID, data.Task.ID)
	}
	item.RenewalNotifMsgData = msgData
}
