package appdeploymentserviceimpl

import (
	"context"
	"fmt"
	"math"
	"strconv"
	"strings"
	"time"

	"github.com/tiendc/gofn"

	"github.com/hivepaas/hivepaas/hivepaas_app/apperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/config"
	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/infra/database"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/githelper"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/vcsurl"
	"github.com/hivepaas/hivepaas/services/git/gitapi"
)

func (s *service) notifyPRForDeploymentResult(
	ctx context.Context,
	db database.IDB,
	data *appDeploymentData,
) error {
	if data == nil || data.App == nil || data.Deployment == nil || data.Deployment.Settings == nil {
		return nil
	}

	deployment := data.Deployment
	settings := deployment.Settings
	if settings.ActiveMethod != base.DeploymentMethodRepo || settings.RepoSource == nil {
		return nil
	}

	repoSource := settings.RepoSource
	pullNumber := extractPRNumber(repoSource.RepoRef, deployment.Trigger)
	if pullNumber <= 0 {
		return nil
	}

	parsedURL, err := vcsurl.Parse(repoSource.RepoURL)
	if err != nil {
		return apperrors.Wrap(err)
	}

	owner := parsedURL.Username
	repo := parsedURL.Name

	message := s.buildPRDeploymentResultMessage(data)
	return s.dispatchPRComment(ctx, db, data, owner, repo, pullNumber, message)
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

func (s *service) buildPRDeploymentResultMessage(data *appDeploymentData) string {
	deployment := data.Deployment
	app := data.App
	isSucceeded := deployment.IsDone()
	scope := app.GetObjectScope()

	dashboardURL := config.Current.DashboardAppDeploymentDetailsURL(scope.GetBaseURLPath(), deployment.ID)
	duration := deployment.GetDuration().Truncate(time.Second).String()

	var commitInfo string
	if deployment.Output != nil {
		shortHash := gofn.Head(deployment.Output.CommitHash, 7) //nolint:mnd
		if shortHash != "" {
			commitInfo = fmt.Sprintf("`%s`", shortHash)
			if deployment.Output.CommitTitle != "" {
				commitInfo += fmt.Sprintf(" - *%s*", deployment.Output.CommitTitle)
			}
			if deployment.Output.CommitAuthor != "" {
				commitInfo += fmt.Sprintf(" (@%s)", deployment.Output.CommitAuthor)
			}
		}
	}
	if commitInfo == "" {
		commitInfo = "*(latest commit)*"
	}

	if isSucceeded {
		previewURL := s.resolveAppPreviewURL(app)
		return buildDeploySucceededPRComment(app.Name, previewURL, commitInfo, duration, dashboardURL)
	}

	var errorText string
	if deployment.Output != nil && len(deployment.Output.Errors) > 0 {
		errorText = strings.Join(deployment.Output.Errors, "\n")
	}
	if errorText == "" {
		errorText = "Deployment task failed during execution. Please inspect build logs for more details."
	}
	return buildDeployFailedPRComment(app.Name, errorText, commitInfo, duration, dashboardURL)
}

func buildDeploySucceededPRComment(
	appName string,
	previewURL string,
	commitInfo string,
	duration string,
	dashboardURL string,
) string {
	var sb strings.Builder
	sb.WriteString("### 🚀 **Preview Deployment Ready!**\n\n")
	sb.WriteString("| Item | Details |\n")
	sb.WriteString("| :--- | :--- |\n")
	if previewURL != "" {
		fmt.Fprintf(&sb, "| **Preview URL** | 🔗 [%s](%s) |\n", previewURL, previewURL)
	}
	fmt.Fprintf(&sb, "| **Application** | `%s` |\n", appName)
	fmt.Fprintf(&sb, "| **Commit** | %s |\n", commitInfo)
	fmt.Fprintf(&sb, "| **Duration** | ⏱️ %s |\n\n", duration)
	fmt.Fprintf(&sb, "👉 [**View Deployment Details on HivePaaS**](%s)", dashboardURL)

	return sb.String()
}

func buildDeployFailedPRComment(
	appName string,
	errorText string,
	commitInfo string,
	duration string,
	dashboardURL string,
) string {
	var sb strings.Builder
	sb.WriteString("### ❌ **Preview Deployment Failed**\n\n")
	fmt.Fprintf(&sb, "The deployment for application `%s` encountered an error:\n\n", appName)
	sb.WriteString("```text\n")
	sb.WriteString(errorText)
	sb.WriteString("\n```\n\n")
	sb.WriteString("| Item | Details |\n")
	sb.WriteString("| :--- | :--- |\n")
	fmt.Fprintf(&sb, "| **Commit** | %s |\n", commitInfo)
	fmt.Fprintf(&sb, "| **Duration** | ⏱️ %s (Failed) |\n\n", duration)
	fmt.Fprintf(&sb, "🔍 [**View Full Build Logs on HivePaaS Dashboard**](%s)", dashboardURL)

	return sb.String()
}

func (s *service) resolveAppPreviewURL(app *entity.App) string {
	if app == nil {
		return ""
	}
	routingSetting := app.GetSettingByType(base.SettingTypeAppRouting)
	if routingSetting == nil {
		return ""
	}
	routing, err := routingSetting.AsAppRoutingSettings()
	if err != nil || routing == nil || !routing.ExposePublicly {
		return ""
	}
	for _, d := range routing.Domains {
		if d.Enabled && d.Domain != "" {
			protocol := "https://"
			if !d.ForceHttps {
				protocol = "http://"
			}
			return protocol + d.Domain
		}
	}
	return ""
}

func (s *service) dispatchPRComment(
	ctx context.Context,
	db database.IDB,
	data *appDeploymentData,
	owner string,
	repo string,
	pullNumber int,
	message string,
) (err error) {
	if message == "" {
		return nil
	}
	repoSource := data.Deployment.Settings.RepoSource

	// 1. Try credentials from deployment repo source
	if repoSource.Credentials.ID != "" {
		credSetting := data.RefObjects.RefSettings[repoSource.Credentials.ID]
		if credSetting == nil {
			credSetting, err = s.settingRepo.GetByID(ctx, db, data.App.GetObjectScope(), "",
				repoSource.Credentials.ID, true)
			if err != nil {
				return apperrors.Wrap(err)
			}
		}
		if credSetting != nil {
			err = gitapi.CreatePullRequestCommentWithRetry(ctx, credSetting, owner, repo, pullNumber, message)
			return apperrors.Wrap(err)
		}
	}

	// 2. Try credentials from webhook trigger
	if data.Deployment.Trigger != nil &&
		data.Deployment.Trigger.Source == base.DeploymentTriggerSourceRepoWebhook &&
		data.Deployment.Trigger.SourceID != "" {
		webhookSetting, err := s.settingRepo.GetByID(ctx, db, nil, "", data.Deployment.Trigger.SourceID, true)
		if err != nil {
			return apperrors.Wrap(err)
		}
		if webhookSetting != nil {
			err = gitapi.CreatePullRequestCommentWithRetry(ctx, webhookSetting, owner, repo, pullNumber, message)
			return apperrors.Wrap(err)
		}
	}

	return nil
}
