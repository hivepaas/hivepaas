package taskschedjobexec

import (
	"context"
	"fmt"
	"time"

	"github.com/tiendc/gofn"

	"github.com/hivepaas/hivepaas/hivepaas_app/config"
	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
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

	isSucceeded := data.Task.IsDone()
	notification, err := e.notificationService.GetNotificationForEvent(ctx, db,
		data.Scope, notifConfig, isSucceeded, data.RefObjects)
	if err != nil {
		return hperrors.Wrap(err)
	}
	if notification == nil {
		return nil
	}

	e.buildNotificationMsgData(data)
	_, err = e.notificationService.NotifyForTaskResult(ctx, db, &notificationservice.TaskResultNotificationReq{
		ActionSucceeded: isSucceeded,
		Scope:           data.Scope,
		RefObjects:      data.RefObjects,
		Notification:    notification,
		TemplateName:    notificationservice.TemplateSchedTaskNotification,
		TemplateData:    data.NotifMsgData,
	})
	if err != nil {
		return hperrors.Wrap(err)
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
			Title: e.notificationService.BuildTitlePrefixForScope(data.Scope) +
				gofn.If(isSucceeded, " Scheduled task succeeded", " Scheduled task failed"),
		},
		Succeeded:    isSucceeded,
		SchedJobName: data.SchedJob.Name,
		StartedAt:    data.Task.StartedAt.Truncate(time.Second),
		Duration:     data.Task.GetDuration().Truncate(time.Millisecond),
		Retries:      data.Task.Config.Retry,
		LastError:    data.Task.GetLastError(),
	}
	msgData.Schedule = describeSchedule(schedJob.Schedule)
	if project := data.Scope.GetProject(); project != nil {
		msgData.ProjectName = project.Name
	}
	if app := data.Scope.GetApp(); app != nil {
		msgData.AppName = app.Name
	}
	msgData.DashboardLink = config.Current().DashboardTaskDetailsURL(data.Scope.GetBaseURLPath(),
		data.Task.ID, data.Scope.ScopeType)

	data.NotifMsgData = msgData
}

// describeSchedule is how a notification names a job's schedule.
func describeSchedule(sched *entity.SchedJobSchedule) string {
	switch {
	case sched == nil:
		return "no schedule (run by hand or by a job sequence)"
	case sched.Interval > 0:
		return fmt.Sprintf("every %v", sched.Interval.String())
	default:
		return fmt.Sprintf("cron expression %v", sched.CronExpr)
	}
}
