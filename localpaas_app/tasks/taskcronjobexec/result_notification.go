package taskcronjobexec

import (
	"context"
	"fmt"

	"github.com/tiendc/gofn"

	"github.com/localpaas/localpaas/localpaas_app/apperrors"
	"github.com/localpaas/localpaas/localpaas_app/base"
	"github.com/localpaas/localpaas/localpaas_app/config"
	"github.com/localpaas/localpaas/localpaas_app/infra/database"
	"github.com/localpaas/localpaas/localpaas_app/service/notificationservice"
)

func (e *Executor) sendNotification(
	ctx context.Context,
	db database.IDB,
	data *taskData,
) (err error) {
	notifConfig := data.CronJob.Notification
	if notifConfig == nil {
		return nil
	}

	var scope *base.SettingScope
	switch {
	case data.App != nil:
		scope = data.App.GetSettingScope()
	case data.Project != nil:
		scope = data.Project.GetSettingScope()
	default:
		scope = base.NewSettingScopeGlobal()
	}

	isSucceeded := data.Task.IsDone()
	notification, err := e.notificationService.GetNotificationForEvent(ctx, db,
		scope, notifConfig, isSucceeded, data.RefObjects)
	if err != nil {
		return apperrors.Wrap(err)
	}
	if notification == nil {
		return nil
	}

	e.buildNotificationMsgData(data)
	_, err = e.notificationService.NotifyForTaskResult(ctx, db, &notificationservice.TaskResultNotificationReq{
		ActionSucceeded: isSucceeded,
		ScopeProject:    data.Project,
		ScopeApp:        data.App,
		RefObjects:      data.RefObjects,
		Notification:    notification,
		TemplateName:    notificationservice.TemplateCronTaskNotification,
		TemplateData:    data.NotifMsgData,
	})
	if err != nil {
		return apperrors.Wrap(err)
	}
	return nil
}

func (e *Executor) buildNotificationMsgData(
	data *taskData,
) {
	isSucceeded := data.Task.IsDone()
	msgData := &notificationservice.TemplateDataCronTask{
		BaseTemplateData: notificationservice.BaseTemplateData{
			Title: e.notificationService.BuildTitlePrefix(data.Project, data.App, nil) +
				gofn.If(isSucceeded, " Scheduled task succeeded", " Scheduled task failed"),
		},
		Succeeded:   isSucceeded,
		CronJobName: data.CronJobSetting.Name,
		CreatedAt:   data.CronJob.Schedule.InitialTime,
		StartedAt:   data.Task.StartedAt,
		Duration:    data.Task.GetDuration(),
		Retries:     data.Task.Config.Retry,
	}
	if data.CronJob.Schedule.Interval > 0 {
		msgData.Schedule = fmt.Sprintf("every %v", data.CronJob.Schedule.Interval.String())
	} else {
		msgData.Schedule = fmt.Sprintf("cron expression %v", data.CronJob.Schedule.CronExpr)
	}
	if data.Project != nil {
		msgData.ProjectName = data.Project.Name
	}
	if data.App != nil {
		msgData.AppName = data.App.Name
	}
	switch {
	case data.App != nil:
		msgData.DashboardLink = config.Current.DashboardAppCronTaskDetailsURL(data.App.ID, data.App.ProjectID,
			data.CronJobSetting.ID, data.Task.ID)
	case data.Project != nil:
		msgData.DashboardLink = config.Current.DashboardProjectCronTaskDetailsURL(data.Project.ID,
			data.CronJobSetting.ID, data.Task.ID)
	default:
		msgData.DashboardLink = config.Current.DashboardGlobalCronTaskDetailsURL(
			data.CronJobSetting.ID, data.Task.ID)
	}
	data.NotifMsgData = msgData
}
