package apppreviewserviceimpl

import (
	"context"
	"fmt"
	"math"
	"strconv"
	"strings"

	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/infra/database"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/githelper"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/vcsurl"
	"github.com/hivepaas/hivepaas/services/git/gitapi"
)

func (s *service) notifyPRForPreviewFailure(
	ctx context.Context,
	db database.IDB,
	data *createPreviewData,
	previewErr error,
) error {
	if data == nil || data.Task == nil || previewErr == nil {
		return nil
	}

	taskArgs := data.Args
	if taskArgs == nil {
		var err error
		taskArgs, err = data.Task.ArgsAsAppPreview()
		if err != nil {
			return hperrors.Wrap(err)
		}
		if taskArgs == nil {
			return nil
		}
		data.Args = taskArgs
	}

	pullNumber := extractPRNumber(taskArgs.RepoRef, taskArgs.Trigger)
	if pullNumber <= 0 {
		return nil
	}

	app := data.App
	if app == nil && taskArgs.ParentApp.ID != "" {
		var err error
		app, err = s.appService.LoadApp(ctx, db, "", taskArgs.ParentApp.ID, true, false)
		if err != nil {
			return hperrors.Wrap(err)
		}
		data.App = app
	}
	if app == nil {
		return nil
	}

	deploymentSetting := app.GetSettingByType(base.SettingTypeAppDeployment)
	if deploymentSetting == nil {
		return nil
	}

	deploymentSettings, err := deploymentSetting.AsAppDeploymentSettings()
	if err != nil {
		return hperrors.Wrap(err)
	}
	if deploymentSettings.ActiveMethod != base.DeploymentMethodRepo ||
		deploymentSettings.RepoSource == nil {
		return nil
	}

	repoSource := deploymentSettings.RepoSource
	parsedURL, err := vcsurl.Parse(repoSource.RepoURL)
	if err != nil {
		return hperrors.Wrap(err)
	}

	owner := parsedURL.Username
	repo := parsedURL.Name

	message := buildPreviewFailedPRComment(app.Name, taskArgs.RepoRef, previewErr)

	return s.dispatchPRComment(ctx, db, data, repoSource, owner, repo, pullNumber, message)
}

func extractPRNumber(repoRef string, trigger *entity.AppDeploymentTrigger) int {
	if repoRef != "" {
		_, prNum, err := githelper.NormalizePullRef(repoRef)
		if err == nil && prNum > 0 && prNum <= math.MaxInt {
			return int(prNum)
		}
	}
	if trigger != nil && strings.HasPrefix(trigger.ChangeID, "pr-") {
		numStr := strings.TrimPrefix(trigger.ChangeID, "pr-")
		num, err := strconv.Atoi(numStr)
		if err == nil && num > 0 {
			return num
		}
	}
	return 0
}

func buildPreviewFailedPRComment(
	appName string,
	targetRef string,
	previewErr error,
) string {
	var errorText string
	if previewErr != nil {
		errorText = previewErr.Error()
	}
	if errorText == "" {
		errorText = "Unknown error occurred while initializing preview environment."
	}

	var sb strings.Builder
	sb.WriteString("### ❌ **Preview Environment Creation Failed**\n\n")
	fmt.Fprintf(&sb, "Failed to initialize preview environment for application `%s`:\n\n", appName)
	sb.WriteString("```text\n")
	sb.WriteString(errorText)
	sb.WriteString("\n```\n\n")
	sb.WriteString("| Item | Details |\n")
	sb.WriteString("| :--- | :--- |\n")
	fmt.Fprintf(&sb, "| **Application** | `%s` |\n", appName)
	if targetRef != "" {
		fmt.Fprintf(&sb, "| **Target Ref** | `%s` |\n", targetRef)
	}
	sb.WriteString("\n👉 *Please check your application preview settings and database dependencies on HivePaaS Dashboard.*")

	return sb.String()
}

func (s *service) dispatchPRComment(
	ctx context.Context,
	db database.IDB,
	data *createPreviewData,
	repoSource *entity.DeploymentRepoSource,
	owner string,
	repo string,
	pullNumber int,
	message string,
) (err error) {
	// 1. Try credentials from deployment repo source
	if repoSource.Credentials.ID != "" {
		credSetting := data.RefObjects.RefSettings[repoSource.Credentials.ID]
		if credSetting == nil {
			credSetting, err = s.settingRepo.GetByID(ctx, db, data.App.GetObjectScope(), "",
				repoSource.Credentials.ID, true)
			if err != nil {
				return hperrors.Wrap(err)
			}
		}
		if credSetting != nil {
			err = gitapi.CreatePullRequestCommentWithRetry(ctx, credSetting, owner, repo, pullNumber, message)
			return hperrors.Wrap(err)
		}
	}

	// 2. Try credentials from webhook trigger
	if data.Args != nil && data.Args.Trigger != nil &&
		data.Args.Trigger.Source == base.DeploymentTriggerSourceRepoWebhook &&
		data.Args.Trigger.SourceID != "" {
		webhookSetting, err := s.settingRepo.GetByID(ctx, db, nil, "", data.Args.Trigger.SourceID, true)
		if err != nil {
			return hperrors.Wrap(err)
		}
		if webhookSetting != nil {
			err = gitapi.CreatePullRequestCommentWithRetry(ctx, webhookSetting, owner, repo, pullNumber, message)
			return hperrors.Wrap(err)
		}
	}

	return nil
}
