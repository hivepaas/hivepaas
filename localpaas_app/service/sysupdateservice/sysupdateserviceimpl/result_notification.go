package sysupdateserviceimpl

import (
	"context"

	"github.com/tiendc/gofn"

	"github.com/localpaas/localpaas/localpaas_app/apperrors"
	"github.com/localpaas/localpaas/localpaas_app/base"
	"github.com/localpaas/localpaas/localpaas_app/config"
	"github.com/localpaas/localpaas/localpaas_app/infra/database"
	"github.com/localpaas/localpaas/localpaas_app/service/notificationservice"
)

func (s *service) notifyForSystemUpdate(
	ctx context.Context,
	db database.IDB,
	data *sysUpdateData,
) (err error) {
	notification, err := s.notificationService.GetDefaultNotification(ctx, db, base.NewSettingScopeGlobal(),
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
		StartedAt:      task.StartedAt,
		Duration:       task.GetDuration(),
		DashboardLink:  config.Current.DashboardTaskDetailsURL(task.ID),
	}
	data.NotifMsgData = msgData
}
