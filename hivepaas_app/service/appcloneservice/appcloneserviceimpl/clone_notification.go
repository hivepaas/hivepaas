package appcloneserviceimpl

import (
	"context"
	"time"

	"github.com/tiendc/gofn"

	"github.com/hivepaas/hivepaas/hivepaas_app/config"
	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/logging"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/safego"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/notificationservice"
)

// cloneReasonMaxLen is how much of a failure's reason a notification carries:
// within what each channel takes in one field - Discord's 1024 characters the
// least. The rest is a click away, in the task's details.
const cloneReasonMaxLen = 1000

// notifyOnCloneEnd tells the targets the clone's settings name how it ended,
// once its task's transaction is over: only then is it known whether the task
// is done, failed for good, or to be tried again - which tells nobody anything.
// A clone with no notification settings - a preview's - tells nobody.
func (s *service) notifyOnCloneEnd(
	ctx context.Context,
	data *appCloneData,
) {
	if data.TaskExecData == nil || data.Task == nil || data.CloneSettings.Notification == nil {
		return
	}
	data.OnPostTx(func() {
		defer safego.Recover("appclone.notifyOnCloneEnd")
		task := data.Task
		if !task.IsDone() && !task.IsFailedCompletely() {
			return
		}
		if err := s.notifyForClone(context.WithoutCancel(ctx), data); err != nil {
			logging.Warnf("app clone: notifying of task %s: %v", task.ID, hperrors.GetErrorDetail(err, ""))
		}
	})
}

func (s *service) notifyForClone(
	ctx context.Context,
	data *appCloneData,
) error {
	srcApp := data.SrcApp
	if srcApp == nil || srcApp.Project == nil {
		return nil
	}
	succeeded := data.Task.IsDone()
	scope := srcApp.GetObjectScope()
	// The targets' own settings - a webhook, a mail sender - are loaded into
	// these, which sending reads them from.
	refObjects := entity.NewRefObjects()
	notification, err := s.notificationService.GetNotificationForEvent(ctx, s.db,
		scope, data.CloneSettings.Notification, succeeded, refObjects)
	if err != nil {
		return hperrors.Wrap(err)
	}
	if notification == nil {
		return nil
	}

	_, err = s.notificationService.NotifyForTaskResult(ctx, s.db, &notificationservice.TaskResultNotificationReq{
		ActionSucceeded: succeeded,
		Scope:           scope,
		RefObjects:      refObjects,
		Notification:    notification,
		TemplateName:    notificationservice.TemplateAppCloneNotification,
		TemplateData:    s.cloneNotifMsgData(data, succeeded),
	})
	if err != nil {
		return hperrors.Wrap(err)
	}
	return nil
}

func (s *service) cloneNotifMsgData(
	data *appCloneData,
	succeeded bool,
) *notificationservice.TemplateDataAppClone {
	srcApp, task := data.SrcApp, data.Task
	msgData := &notificationservice.TemplateDataAppClone{
		BaseTemplateData: notificationservice.BaseTemplateData{
			Title: s.notificationService.BuildTitlePrefix(srcApp.Project, srcApp, nil) +
				gofn.If(succeeded, " Clone succeeded", " Clone failed"),
		},
		ProjectName: srcApp.Project.Name,
		AppName:     srcApp.Name,
		CloneName:   data.CloneSettings.TargetName,
		CloneEnv:    data.CloneSettings.TargetEnv,
		Succeeded:   succeeded,
		StartedAt:   task.StartedAt.Truncate(time.Second),
		Duration:    task.GetDuration().Truncate(time.Millisecond),
		DashboardLink: config.Current().DashboardTaskDetailsURL(srcApp.GetObjectScope().GetBaseURLPath(),
			task.ID, srcApp.GetObjectScope().ScopeType),
	}
	// What the clone was made as, when it got that far: the settings may leave
	// its name and environment to be the source's.
	if dest := data.DestApp; dest != nil {
		msgData.CloneName = dest.Name
		if dest.ProjectEnv != nil {
			msgData.CloneEnv = dest.ProjectEnv.Name
		}
	}
	if !succeeded {
		reason := []rune(task.GetLastError())
		if len(reason) > cloneReasonMaxLen {
			reason = append(reason[:cloneReasonMaxLen-1], '…')
		}
		msgData.Reason = string(reason)
	}
	return msgData
}
