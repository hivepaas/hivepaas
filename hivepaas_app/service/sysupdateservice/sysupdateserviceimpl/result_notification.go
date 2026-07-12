package sysupdateserviceimpl

import (
	"context"
	"time"

	"github.com/tiendc/gofn"

	"github.com/hivepaas/hivepaas/hivepaas_app/apperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/config"
	"github.com/hivepaas/hivepaas/hivepaas_app/infra/database"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/notificationservice"
)

func (s *service) notifyForSystemUpdate(
	ctx context.Context,
	db database.IDB,
	data *sysUpdateData,
) (err error) {
	notification, err := s.notificationService.GetDefaultNotification(ctx, db, base.NewObjectScopeGlobal(),
		data.RefObjects, false)
	if err != nil {
		return apperrors.Wrap(err)
	}
	if notification == nil {
		return nil
	}

	s.buildSystemUpdateNotifMsgData(data)
	_, err = s.notificationService.NotifyForTaskResult(ctx, db, &notificationservice.TaskResultNotificationReq{
		ActionSucceeded: data.Task.IsDone(),
		RefObjects:      data.RefObjects,
		Notification:    notification,
		TemplateName:    notificationservice.TemplateSystemUpdateNotification,
		TemplateData:    data.NotifMsgData,
	})
	if err != nil {
		return apperrors.Wrap(err)
	}
	return nil
}

func (s *service) buildSystemUpdateNotifMsgData(
	data *sysUpdateData,
) {
	task := data.Task
	args := gofn.Must(task.ArgsAsSystemUpdate())
	isSucceeded := task.IsDone()
	msgData := &notificationservice.TemplateDataSystemUpdate{
		BaseTemplateData: notificationservice.BaseTemplateData{
			Title: gofn.If(isSucceeded, "System update succeeded", "System update failed"),
		},
		CurrentVersion: args.CurrentVersion.AppVersion,
		TargetVersion:  args.TargetVersion.AppVersion,
		Succeeded:      isSucceeded,
		StartedAt:      task.StartedAt.Truncate(time.Second),
		Duration:       task.GetDuration().Truncate(time.Millisecond),
		DashboardLink:  config.Current.DashboardTaskDetailsURL(task.ID),
	}
	data.NotifMsgData = msgData
}
