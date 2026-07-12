package taskschedjobexec

import (
	"context"
	"fmt"
	"time"

	"github.com/tiendc/gofn"

	"github.com/hivepaas/hivepaas/hivepaas_app/apperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/config"
	"github.com/hivepaas/hivepaas/hivepaas_app/infra/database"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/notificationservice"
)

func (e *Executor) sendNotification(
	ctx context.Context,
	db database.IDB,
	data *taskData,
) (err error) {
	schedJob := data.SchedJob.MustAsSchedJob()
	notifConfig := schedJob.Notification
	if notifConfig == nil {
		return nil
	}

	var scope *base.ObjectScope
	switch {
	case data.App != nil:
		scope = data.App.GetObjectScope()
	case data.Project != nil:
		scope = data.Project.GetObjectScope()
	default:
		scope = base.NewObjectScopeGlobal()
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
		TemplateName:    notificationservice.TemplateSchedTaskNotification,
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
	schedJob := data.SchedJob.MustAsSchedJob()
	isSucceeded := data.Task.IsDone()
	msgData := &notificationservice.TemplateDataSchedTask{
		BaseTemplateData: notificationservice.BaseTemplateData{
			Title: e.notificationService.BuildTitlePrefix(data.Project, data.App, nil) +
				gofn.If(isSucceeded, " Scheduled task succeeded", " Scheduled task failed"),
		},
		Succeeded:    isSucceeded,
		SchedJobName: data.SchedJob.Name,
		StartedAt:    data.Task.StartedAt.Truncate(time.Second),
		Duration:     data.Task.GetDuration().Truncate(time.Millisecond),
		Retries:      data.Task.Config.Retry,
	}
	if schedJob.Schedule.Interval > 0 {
		msgData.Schedule = fmt.Sprintf("every %v", schedJob.Schedule.Interval.String())
	} else {
		msgData.Schedule = fmt.Sprintf("cron expression %v", schedJob.Schedule.CronExpr)
	}
	if data.Project != nil {
		msgData.ProjectName = data.Project.Name
	}
	if data.App != nil {
		msgData.AppName = data.App.Name
	}
	switch {
	case data.App != nil:
		msgData.DashboardLink = config.Current.DashboardAppSchedTaskDetailsURL(data.App.ProjectID, data.App.ID,
			data.SchedJob.ID, data.Task.ID)
	case data.Project != nil:
		msgData.DashboardLink = config.Current.DashboardProjectSchedTaskDetailsURL(data.Project.ID,
			data.SchedJob.ID, data.Task.ID)
	default:
		msgData.DashboardLink = config.Current.DashboardGlobalSchedTaskDetailsURL(
			data.SchedJob.ID, data.Task.ID)
	}
	data.NotifMsgData = msgData
}
