package apppreviewserviceimpl

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/config"
	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/logging"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/safego"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/notificationservice"
)

// previewPlanLimit is the plan that refused a preview, and how many previews it
// runs at once.
type previewPlanLimit struct {
	PlanName string
	Max      int
}

// Sentence is the refusal as the pull request and the app's notification target
// are told it: what the plan allows, and what to do.
func (l *previewPlanLimit) Sentence() string {
	switch {
	case l.Max <= 0:
		// Only a key can say so, and none is sold that does; but "close one"
		// would send the reader looking for a preview there cannot be.
		return fmt.Sprintf("%s runs no previews: upgrade.", l.PlanName)
	case l.Max == 1:
		return fmt.Sprintf("%s runs 1 preview at a time: close it, or upgrade.", l.PlanName)
	}
	return fmt.Sprintf("%s runs %d previews at a time: close one, or upgrade.", l.PlanName, l.Max)
}

// previewPlanLimitOf is the plan's limit when previewErr is the plan refusing the
// preview, and nil for any other failure. It is read from the license in force,
// which is what the refusal was decided on a moment before. When that cannot be
// read the failure is reported as any other.
func (s *service) previewPlanLimitOf(ctx context.Context, previewErr error) *previewPlanLimit {
	if !errors.Is(previewErr, hperrors.ErrPlanLimitReached) {
		return nil
	}
	state, err := s.licenseService.Current(ctx)
	if err != nil {
		logging.Warnf("app preview: reading the license a preview was refused by: %v",
			hperrors.GetErrorDetail(err, ""))
		return nil
	}
	maxCount, limited := state.Limits.Max(base.LicenseLimitPreviews)
	if !limited {
		// A key that lifts the limit was installed since: there is no limit to
		// name.
		return nil
	}
	return &previewPlanLimit{PlanName: state.Plan.DisplayName(), Max: maxCount}
}

// notifyPlanLimitRefused tells the app's notification target that the plan
// refused a preview: when a pull request comment asks for one, nobody is looking
// at a screen. It is told as a preview's deployment would tell it - the target
// the app's deployment settings name for a failure, and the deployment
// notification, failed, with the plan's sentence as the reason - which every
// channel already has a template for.
//
// It is sent once the task's transaction is over: the plan-limit lock the
// transaction holds is not kept while channels are called.
func (s *service) notifyPlanLimitRefused(
	ctx context.Context,
	data *createPreviewData,
	planLimit *previewPlanLimit,
) {
	if planLimit == nil || data == nil || data.TaskExecData == nil || data.Task == nil {
		return
	}
	app := data.App
	if app == nil || app.Project == nil || app.ProjectEnv == nil {
		return
	}
	deploymentSettings := appDeploymentSettings(app)
	if deploymentSettings == nil || deploymentSettings.Notification == nil {
		return
	}

	data.OnPostTx(func() {
		defer safego.Recover("apppreview.notifyPlanLimitRefused")
		err := s.notifyForPlanLimit(ctx, data, deploymentSettings, planLimit)
		if err != nil {
			logging.Warnf("app preview: notifying of task %s refused by the plan: %v", data.Task.ID,
				hperrors.GetErrorDetail(err, ""))
		}
	})
}

func (s *service) notifyForPlanLimit(
	ctx context.Context,
	data *createPreviewData,
	deploymentSettings *entity.AppDeploymentSettings,
	planLimit *previewPlanLimit,
) error {
	scope := data.App.GetObjectScope()
	// The targets' own settings - a webhook, a mail sender - are loaded into
	// these, which sending reads them from.
	refObjects := entity.NewRefObjects()
	notification, err := s.notificationService.GetNotificationForEvent(ctx, s.db,
		scope, deploymentSettings.Notification, false, refObjects)
	if err != nil {
		return hperrors.Wrap(err)
	}
	if notification == nil {
		return nil
	}

	_, err = s.notificationService.NotifyForTaskResult(ctx, s.db, &notificationservice.TaskResultNotificationReq{
		ActionSucceeded: false,
		Scope:           scope,
		RefObjects:      refObjects,
		Notification:    notification,
		TemplateName:    notificationservice.TemplateAppDeploymentNotification,
		TemplateData:    s.planLimitNotifMsgData(data, deploymentSettings, planLimit),
	})
	if err != nil {
		return hperrors.Wrap(err)
	}
	return nil
}

func (s *service) planLimitNotifMsgData(
	data *createPreviewData,
	deploymentSettings *entity.AppDeploymentSettings,
	planLimit *previewPlanLimit,
) *notificationservice.TemplateDataAppDeployment {
	app, task := data.App, data.Task
	msgData := &notificationservice.TemplateDataAppDeployment{
		BaseTemplateData: notificationservice.BaseTemplateData{
			Title: s.notificationService.BuildTitlePrefix(app.Project, app, nil) + " Preview refused",
		},
		ProjectName:   app.Project.Name,
		AppName:       app.Name,
		Succeeded:     false,
		Method:        deploymentSettings.ActiveMethod,
		StartedAt:     task.StartedAt.Truncate(time.Second),
		Duration:      task.GetDuration().Truncate(time.Millisecond),
		DashboardLink: config.Current().DashboardAppPreviewsURL(app.GetObjectScope().GetBaseURLPath()),
		Reason:        planLimit.Sentence(),
	}
	if deploymentSettings.RepoSource != nil {
		msgData.RepoURL = deploymentSettings.RepoSource.RepoURL
	}
	// The ref the preview was asked for, not the app's own branch.
	if data.Args != nil {
		msgData.RepoRef = data.Args.RepoRef
	}
	return msgData
}

// appDeploymentSettings is the app's deployment settings, or nil when it has none
// or they do not parse.
func appDeploymentSettings(app *entity.App) *entity.AppDeploymentSettings {
	setting := app.GetSettingByType(base.SettingTypeAppDeployment)
	if setting == nil {
		return nil
	}
	deploymentSettings, err := setting.AsAppDeploymentSettings()
	if err != nil {
		return nil
	}
	return deploymentSettings
}
