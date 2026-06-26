package webhookuc

import (
	"context"
	"strconv"
	"strings"
	"time"

	"github.com/gitsight/go-vcsurl"

	"github.com/localpaas/localpaas/localpaas_app/apperrors"
	"github.com/localpaas/localpaas/localpaas_app/base"
	"github.com/localpaas/localpaas/localpaas_app/entity"
	"github.com/localpaas/localpaas/localpaas_app/infra/database"
	"github.com/localpaas/localpaas/localpaas_app/pkg/bunex"
	"github.com/localpaas/localpaas/localpaas_app/pkg/githelper"
	"github.com/localpaas/localpaas/localpaas_app/pkg/timeutil"
	"github.com/localpaas/localpaas/localpaas_app/service/apppreviewservice"
	"github.com/localpaas/localpaas/localpaas_app/service/appservice"
)

const (
	previewCmdDeploy             = "deploy"
	previewCmdDeployArgNoStart   = "nostart"
	previewCmdDeployArgNoWait    = "nowait"
	previewCmdDeployArgSubdomain = "subdomain"

	previewCmdCancel = "cancel"
)

const (
	deployDelayDefault = 30 * time.Second
)

type repoPRCommentEventData struct {
	RepoID      string
	RepoURL     string
	PRNumber    int64
	CommentBody string
	Branch      string

	// Parsed command data
	previewCmd             string
	previewDeployNoStart   bool
	previewDeployNoWait    bool
	previewDeploySubdomain string
}

func (uc *UC) processWebhookEventPRComment(
	ctx context.Context,
	db database.Tx,
	prCommentEvent *repoPRCommentEventData,
	data *handleRepoWebhookData,
	persistingData *appservice.PersistingAppData,
) (err error) {
	parsedURL, err := vcsurl.Parse(prCommentEvent.RepoURL)
	if err != nil {
		return apperrors.New(err)
	}
	prCommentEvent.RepoID = parsedURL.ID

	success, err := uc.parsePRCommentCommand(prCommentEvent)
	if err != nil {
		return apperrors.New(err)
	}
	if !success {
		return nil
	}

	var appListOpts []bunex.SelectQueryOption
	switch prCommentEvent.previewCmd {
	case previewCmdDeploy:
		// Load main apps to create previews
		appListOpts = append(appListOpts, bunex.SelectWhere("app.parent_id IS NULL"))
	case previewCmdCancel:
		// Load preview apps to delete
		appListOpts = append(appListOpts, bunex.SelectWhere("app.parent_id IS NOT NULL"))
	}

	apps, err := uc.findAppsMatchingRepository(ctx, db, prCommentEvent.RepoID, "", appListOpts...)
	if err != nil {
		return apperrors.New(err)
	}
	for _, app := range apps {
		switch prCommentEvent.previewCmd {
		case previewCmdDeploy:
			err = uc.createAppPreviewByCommentEvent(ctx, db, app, prCommentEvent, data, persistingData)
		case previewCmdCancel:
			err = uc.deleteAppPreviewByCommentEvent(ctx, db, app, prCommentEvent, data, persistingData)
		}
		if err != nil {
			return apperrors.New(err)
		}
	}
	return nil
}

func (uc *UC) createAppPreviewByCommentEvent(
	ctx context.Context,
	db database.Tx,
	app *entity.App,
	commentEvent *repoPRCommentEventData,
	data *handleRepoWebhookData,
	persistingData *appservice.PersistingAppData,
) error {
	shouldDeploy, err := uc.shouldCreateAppPreviewByCommentEvent(ctx, db, app, commentEvent)
	if err != nil {
		return apperrors.New(err)
	}
	if !shouldDeploy {
		return nil
	}

	var repoRef string
	prNumberStr := strconv.FormatInt(commentEvent.PRNumber, 10)
	webhook := data.WebhookSetting.MustAsRepoWebhook()
	switch webhook.Kind {
	case base.WebhookKindGithub:
		repoRef = "pull/" + prNumberStr + "/head"
	case base.WebhookKindGitlab:
		repoRef = "merge-requests/" + prNumberStr + "/head"
	case base.WebhookKindGitea, base.WebhookKindGogs:
		repoRef = "pull/" + prNumberStr + "/head"
	case base.WebhookKindBitbucket:
		if commentEvent.Branch != "" {
			repoRef = commentEvent.Branch
		}
	default:
		return apperrors.New(apperrors.ErrWebhookTypeUnsupported).WithParam("Type", webhook.Kind)
	}

	createResp, err := uc.appPreviewService.CreatePreview(ctx, db, &apppreviewservice.CreatePreviewReq{
		ProjectID:       app.ProjectID,
		AppID:           app.ID,
		RepoRef:         repoRef,
		NoStart:         commentEvent.previewDeployNoStart,
		CustomSubdomain: commentEvent.previewDeploySubdomain,
		OnInitDeployment: func(deployment *entity.Deployment) error {
			deployment.Trigger = &entity.AppDeploymentTrigger{
				Source:   base.DeploymentTriggerSourceRepoWebhook,
				SourceID: data.WebhookSetting.ID,
				ChangeID: "pr-" + prNumberStr,
			}
			return nil
		},
		OnDeploymentTask: func(task *entity.Task) error {
			if !commentEvent.previewDeployNoWait {
				task.RunAt = timeutil.NowUTC().Add(deployDelayDefault)
			}
			return nil
		},
	})
	if createResp != nil && createResp.OnCleanup != nil {
		defer func() {
			_ = createResp.OnCleanup(err)
		}()
	}
	if err != nil {
		return apperrors.New(err)
	}

	if createResp.DeploymentTask != nil {
		persistingData.UpsertingTasks = append(persistingData.UpsertingTasks, createResp.DeploymentTask)
	}
	return nil
}

//nolint:unparam
func (uc *UC) shouldCreateAppPreviewByCommentEvent(
	_ context.Context,
	_ database.Tx,
	app *entity.App,
	_ *repoPRCommentEventData,
) (bool, error) {
	// The app is already a preview app, skips it
	if app.ParentID != "" {
		return false, nil
	}

	// TODO: Load preview settings and check them

	return true, nil
}

func (uc *UC) deleteAppPreviewByCommentEvent(
	ctx context.Context,
	db database.Tx,
	app *entity.App,
	commentEvent *repoPRCommentEventData,
	data *handleRepoWebhookData,
	persistingData *appservice.PersistingAppData,
) error {
	shouldDelete, err := uc.shouldDeleteAppPreviewByCommentEvent(ctx, db, app, commentEvent)
	if err != nil {
		return apperrors.New(err)
	}
	if !shouldDelete {
		return nil
	}

	webhook := data.WebhookSetting.MustAsRepoWebhook()
	var expectedRef string
	prNumberStr := strconv.FormatInt(commentEvent.PRNumber, 10)

	switch webhook.Kind {
	case base.WebhookKindGithub:
		expectedRef = "refs/pull/" + prNumberStr + "/head"
	case base.WebhookKindGitlab:
		expectedRef = "refs/merge-requests/" + prNumberStr + "/head"
	case base.WebhookKindGitea, base.WebhookKindGogs:
		expectedRef = "refs/pull/" + prNumberStr + "/head"
	case base.WebhookKindBitbucket:
		if commentEvent.Branch != "" {
			expectedRef = string(githelper.NormalizeRepoRef(commentEvent.Branch))
		}
	}

	if expectedRef == "" {
		return nil
	}

	deploymentSetting := app.GetSettingByType(base.SettingTypeAppDeployment)
	if deploymentSetting == nil {
		return nil
	}
	deploymentSettings, err := deploymentSetting.AsAppDeploymentSettings()
	if err != nil {
		return apperrors.New(err)
	}

	if deploymentSettings.RepoSource != nil && deploymentSettings.RepoSource.RepoRef == expectedRef {
		if err = uc.appService.DeleteApp(ctx, db, app); err != nil {
			return apperrors.New(err)
		}
		app.DeletedAt = time.Now()
		app.UpdateVer++
		persistingData.UpsertingApps = append(persistingData.UpsertingApps, app)
	}

	return nil
}

//nolint:unparam
func (uc *UC) shouldDeleteAppPreviewByCommentEvent(
	_ context.Context,
	_ database.Tx,
	app *entity.App,
	_ *repoPRCommentEventData,
) (bool, error) {
	// The app is not a preview, skip it
	if app.ParentID == "" {
		return false, nil
	}

	// TODO: Load preview settings and check them

	return true, nil
}

func (uc *UC) parsePRCommentCommand(
	commentEvent *repoPRCommentEventData,
) (bool, error) {
	var firstValidLine string
	for _, line := range strings.Split(commentEvent.CommentBody, "\n") {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		firstValidLine = line
		break
	}

	if !strings.HasPrefix(firstValidLine, "/localpaas") {
		return false, nil
	}

	fields := strings.Fields(firstValidLine)
	if len(fields) <= 1 {
		return false, nil
	}

	for _, field := range fields[1:] {
		k, v, _ := strings.Cut(field, "=")
		switch {
		case k == previewCmdDeploy || k == previewCmdCancel:
			commentEvent.previewCmd = k
		case (k == previewCmdDeployArgNoStart || k == "no-start") && commentEvent.previewCmd == previewCmdDeploy:
			if v == "" {
				commentEvent.previewDeployNoStart = true
				continue // continue for-loop
			}
			boolVal, err := strconv.ParseBool(v)
			if err != nil {
				return false, apperrors.New(err)
			}
			commentEvent.previewDeployNoStart = boolVal
		case (k == previewCmdDeployArgNoWait || k == "no-wait") && commentEvent.previewCmd == previewCmdDeploy:
			if v == "" {
				commentEvent.previewDeployNoWait = true
				continue // continue for-loop
			}
			boolVal, err := strconv.ParseBool(v)
			if err != nil {
				return false, apperrors.New(err)
			}
			commentEvent.previewDeployNoWait = boolVal
		case k == previewCmdDeployArgSubdomain && commentEvent.previewCmd == previewCmdDeploy:
			commentEvent.previewDeploySubdomain = v
		}
	}

	return commentEvent.previewCmd != "", nil
}
