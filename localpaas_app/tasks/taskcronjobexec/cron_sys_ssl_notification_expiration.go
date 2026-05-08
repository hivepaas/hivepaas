package taskcronjobexec

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

func (e *Executor) sslNotifyOfExpiration(
	ctx context.Context,
	db database.IDB,
	item *sslRenewalTaskItem,
	data *sslRenewalTaskData,
) (err error) {
	isSucceeded := false
	notification, err := e.sslGetNotification(ctx, db, item.Setting, isSucceeded, data)
	if err != nil {
		return apperrors.Wrap(err)
	}
	if notification == nil {
		return nil
	}

	e.sslBuildExpiringNotificationMsgData(item, data)
	_, err = e.notificationService.NotifyForTaskResult(ctx, db, &notificationservice.TaskResultNotificationReq{
		ActionSucceeded: isSucceeded,
		ScopeProject:    data.Project,
		ScopeApp:        data.App,
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

func (e *Executor) sslBuildExpiringNotificationMsgData(
	item *sslRenewalTaskItem,
	data *sslRenewalTaskData,
) {
	ssl := item.Setting.MustAsSSLCert()
	msgData := &notificationservice.TemplateDataSSLExpiring{
		BaseTemplateData: notificationservice.BaseTemplateData{
			Title: e.notificationService.BuildTitlePrefix(data.Project, data.App, nil) +
				fmt.Sprintf(" Your SSL expiring in %v", item.ExpiringNotifMsgData.ExpireIn),
		},
		SSLName:   item.Setting.Name,
		SSLType:   string(ssl.CertType),
		Domain:    ssl.Domain,
		CreatedAt: item.Setting.CreatedAt,
		ExpireAt:  ssl.ExpireAt,
		ExpireIn:  timeutil.Duration(ssl.ExpireAt.Sub(timeutil.NowUTC()).Truncate(time.Hour)),
	}
	project := item.Setting.BelongToProject
	app := item.Setting.BelongToApp
	if project != nil {
		msgData.ProjectName = project.Name
	}
	if app != nil {
		msgData.AppName = app.Name
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
	item.ExpiringNotifMsgData = msgData
}
