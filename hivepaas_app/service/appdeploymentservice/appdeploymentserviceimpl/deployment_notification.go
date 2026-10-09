package appdeploymentserviceimpl

import (
	"context"
	"strings"
	"time"

	"github.com/tiendc/gofn"

	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/config"
	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/infra/database"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/notificationservice"
)

func (s *service) notifyForDeployment(
	ctx context.Context,
	db database.IDB,
	data *appDeploymentData,
) (err error) {
	if data == nil || data.App == nil || data.App.ID == "" {
		return nil
	}
	if data.Deployment == nil || data.Deployment.Settings == nil {
		return nil
	}

	// Reload app to verify it hasn't been soft-deleted or removed during deployment
	err = s.reloadApp(ctx, db, true, true, data)
	if err != nil {
		return hperrors.Wrap(err)
	}

	notifConfig := data.Deployment.Settings.Notification
	if notifConfig == nil {
		return nil
	}

	notification, err := s.notificationService.GetNotificationForEvent(ctx, db,
		data.App.GetObjectScope(), notifConfig, data.Deployment.IsDone(), data.RefObjects)
	if err != nil {
		return hperrors.Wrap(err)
	}
	if notification == nil {
		return nil
	}

	s.buildDeploymentNotifMsgData(data)
	_, err = s.notificationService.NotifyForTaskResult(ctx, db, &notificationservice.TaskResultNotificationReq{
		ActionSucceeded: data.Deployment.IsDone(),
		Scope:           data.App.GetObjectScope(),
		RefObjects:      data.RefObjects,
		Notification:    notification,
		TemplateName:    notificationservice.TemplateAppDeploymentNotification,
		TemplateData:    data.NotifMsgData,
	})
	if err != nil {
		return hperrors.Wrap(err)
	}
	return nil
}

func (s *service) buildDeploymentNotifMsgData(
	data *appDeploymentData,
) {
	scope := data.App.GetObjectScope()
	deployment := data.Deployment
	isSucceeded := deployment.IsDone()
	msgData := &notificationservice.TemplateDataAppDeployment{
		BaseTemplateData: notificationservice.BaseTemplateData{
			Title: s.notificationService.BuildTitlePrefix(data.App.Project, data.App, nil) +
				gofn.If(isSucceeded, " Deployment succeeded", " Deployment failed"),
		},
		ProjectName:   data.App.Project.Name,
		AppName:       data.App.Name,
		Succeeded:     isSucceeded,
		Method:        deployment.Settings.ActiveMethod,
		StartedAt:     deployment.StartedAt.Truncate(time.Second),
		Duration:      deployment.GetDuration().Truncate(time.Millisecond),
		DashboardLink: config.Current().DashboardAppDeploymentDetailsURL(scope.GetBaseURLPath(), deployment.ID),
	}
	if !isSucceeded {
		msgData.Reason = failureReason(deployment)
	}
	data.NotifMsgData = msgData

	switch deployment.Settings.ActiveMethod {
	case base.DeploymentMethodRepo:
		msgData.RepoURL = deployment.Settings.RepoSource.RepoURL
		msgData.RepoRef = deployment.Settings.RepoSource.RepoRef
		if deployment.Output != nil {
			msgData.CommitMsg = deployment.Output.CommitTitle
			msgData.CommitAuthor = deployment.Output.CommitAuthor
		}
	case base.DeploymentMethodImage:
		msgData.Image = deployment.Settings.ImageSource.Image
	case base.DeploymentMethodFunction:
		if source := deployment.Settings.FunctionSource; source != nil && source.Code.Repo != nil {
			msgData.RepoURL = source.Code.Repo.RepoURL
			msgData.RepoRef = source.Code.Repo.RepoRef
			if deployment.Output != nil {
				msgData.CommitMsg = deployment.Output.CommitTitle
				msgData.CommitAuthor = deployment.Output.CommitAuthor
			}
		}
	}
}

// notifReasonMaxLen is how much of a failure's reason a notification carries:
// within what each channel takes in one field - Discord's 1024 characters the
// least. The rest is a click away, in the deployment's details.
const notifReasonMaxLen = 1000

// failureReason is why a deployment failed, as its details say it, cut to what
// a notification carries.
func failureReason(deployment *entity.Deployment) string {
	if deployment.Output == nil {
		return ""
	}
	reason := []rune(strings.Join(deployment.Output.Errors, "\n"))
	if len(reason) > notifReasonMaxLen {
		return string(reason[:notifReasonMaxLen-1]) + "…"
	}
	return string(reason)
}
