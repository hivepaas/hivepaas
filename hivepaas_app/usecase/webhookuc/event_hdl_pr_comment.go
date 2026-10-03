package webhookuc

import (
	"context"
	"net/url"
	"strconv"
	"strings"
	"sync"

	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/config"
	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/infra/database"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/bunex"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/githelper"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/logging"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/projecthelper"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/safego"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/vcsurl"
)

const (
	previewCmdDeploy             = "deploy"
	previewCmdDeployArgNoStart   = "nostart"
	previewCmdDeployArgNoWait    = "nowait"
	previewCmdDeployArgCloneDb   = "clonedb"
	previewCmdDeployArgNoCloneDb = "noclonedb"
	previewCmdDeployArgSubdomain = "subdomain"

	previewCmdCancel = "cancel"
)

type repoPRCommentEventData struct {
	RepoURL     string
	PRNumber    int64
	CommentBody string
	Branch      string
	// Author is who wrote the comment: only someone who may write to the
	// repository runs commands.
	Author prAuthor

	// Parsed command data
	previewCmd             string
	previewDeployNoStart   bool
	previewDeployNoWait    bool
	previewDeployCloneDB   bool
	previewDeployNoCloneDB bool
	previewDeploySubdomain string
}

func (uc *UC) processWebhookEventPRComment(
	ctx context.Context,
	db database.IDB,
	prCommentEvent *repoPRCommentEventData,
	data *handleRepoWebhookData,
) (err error) {
	parsedURL, err := vcsurl.Parse(prCommentEvent.RepoURL)
	if err != nil {
		return hperrors.Wrap(err)
	}

	isHivepaasCmd, success, rawCmd, _ := uc.parsePRCommentCommand(prCommentEvent)
	if !isHivepaasCmd {
		return nil
	}

	var repoRef string
	webhook := data.WebhookSetting.MustAsRepoWebhook()
	if webhook.Kind == base.WebhookKindBitbucket && prCommentEvent.Branch != "" {
		repoRef = string(githelper.NormalizeRepoRef(prCommentEvent.Branch))
	}
	if repoRef == "" {
		repoRef, _ = githelper.GetPullNumberRef(prCommentEvent.PRNumber, base.GitSource(webhook.Kind))
	}
	if repoRef == "" {
		return nil
	}

	apps, err := uc.appService.FindAppsMatchingRepository(ctx, db, parsedURL.ID, "",
		bunex.SelectExcludeColumns(entity.AppDefaultExcludeColumns...),
		// If cancels, load preview apps to delete
		bunex.SelectWhereIf(prCommentEvent.previewCmd == previewCmdCancel, "app.parent_id IS NOT NULL"),
	)
	if err != nil {
		return hperrors.Wrap(err)
	}

	var firstApp *entity.App
	if len(apps) > 0 {
		firstApp = apps[0]
	}

	// 0. Only someone who may write to the repository runs commands: a preview
	// runs the pull request's code with the app's env vars.
	if !uc.prAuthorAllowed(ctx, db, prCommentEvent.RepoURL, &prCommentEvent.Author, data, apps) {
		logging.Warnf("webhook: %s may not write to %s: its command on pull request %d is refused",
			prCommentEvent.Author.Login, prCommentEvent.RepoURL, prCommentEvent.PRNumber)
		_ = uc.sendPRComment(ctx, db, prCommentEvent, data, firstApp, buildAuthorNotAllowedComment())
		return nil
	}

	// 1. If command is invalid, notify the user with usage instructions
	if !success {
		_ = uc.sendPRComment(ctx, db, prCommentEvent, data, firstApp, buildInvalidCommandComment(rawCmd))
		return nil
	}

	// 2. If no apps match the repository
	if len(apps) == 0 {
		switch prCommentEvent.previewCmd {
		case previewCmdCancel:
			_ = uc.sendPRComment(ctx, db, prCommentEvent, data, nil, buildNoActivePreviewComment())
		case previewCmdDeploy:
			_ = uc.sendPRComment(ctx, db, prCommentEvent, data, nil, buildAppNotFoundComment(parsedURL.Name))
		}
		return nil
	}

	// 3. Process valid commands
	var wg sync.WaitGroup
	for _, app := range apps {
		wg.Go(func() {
			defer safego.Recover("webhook.prComment.handleCommand")
			switch prCommentEvent.previewCmd {
			case previewCmdDeploy:
				uc.handlePRCommentDeploy(ctx, db, app, prCommentEvent, repoRef, data)
			case previewCmdCancel:
				uc.handlePRCommentCancel(ctx, db, app, prCommentEvent, repoRef, data)
			}
		})
	}
	wg.Wait()

	return nil
}

func (uc *UC) handlePRCommentDeploy(
	ctx context.Context,
	db database.IDB,
	app *entity.App,
	prCommentEvent *repoPRCommentEventData,
	repoRef string,
	data *handleRepoWebhookData,
) {
	previewSettings, refusal := uc.prCommentsGate(ctx, db, app)
	if refusal != "" {
		_ = uc.sendPRComment(ctx, db, prCommentEvent, data, app, refusal)
		return
	}

	if !app.IsPreviewApp() {
		err := uc.createAppPreview(ctx, app, prCommentEvent, repoRef, data.WebhookSetting.ID, previewSettings)
		if err != nil {
			_ = uc.sendPRComment(ctx, db, prCommentEvent, data, app, buildDeployFailedComment(app.Name, err))
			return
		}

		cloneDBApps := prCommentEvent.previewDeployCloneDB
		if !prCommentEvent.previewDeployNoCloneDB {
			cloneDBApps = cloneDBApps || (previewSettings.AutoCloneApps && len(previewSettings.AppsToClone) > 0)
		}
		_ = uc.sendPRComment(ctx, db, prCommentEvent, data, app,
			buildDeployPreviewComment(cloneDBApps, uc.withheldSecretsNote(ctx, db, app)))
	} else {
		// TODO: find the SHA of the head commit of the PR (change id)
		err := uc.createAppDeployment(ctx, app, "", data.WebhookSetting.ID)
		if err != nil {
			_ = uc.sendPRComment(ctx, db, prCommentEvent, data, app, buildDeployFailedComment(app.Name, err))
			return
		}
		_ = uc.sendPRComment(ctx, db, prCommentEvent, data, app,
			buildDeployPreviewComment(true, uc.withheldSecretsNote(ctx, db, app)))
	}
}

func (uc *UC) handlePRCommentCancel(
	ctx context.Context,
	db database.IDB,
	app *entity.App,
	prCommentEvent *repoPRCommentEventData,
	repoRef string,
	data *handleRepoWebhookData,
) {
	if _, refusal := uc.prCommentsGate(ctx, db, app); refusal != "" {
		_ = uc.sendPRComment(ctx, db, prCommentEvent, data, app, refusal)
		return
	}

	err := uc.deleteAppPreview(ctx, app, repoRef)
	if err != nil {
		_ = uc.sendPRComment(ctx, db, prCommentEvent, data, app, buildDeployFailedComment(app.Name, err))
		return
	}
	_ = uc.sendPRComment(ctx, db, prCommentEvent, data, app, buildCancelPreviewComment())
}

// prCommentsGate decides whether a pull request's comments may run commands on
// an app, by the preview settings of the app the previews are of - the app's
// own, or a preview's parent's. It answers those settings, and the reply that
// refuses the command, or "" when it may run.
func (uc *UC) prCommentsGate(
	ctx context.Context,
	db database.IDB,
	app *entity.App,
) (*entity.AppFeaturePreviewSettings, string) {
	ownerID := app.ID
	if app.IsPreviewApp() {
		ownerID = app.ParentID
	}
	owner, featureSettings, err := uc.appService.LoadAppWithFeatureSettings(ctx, db, app.ProjectID, ownerID,
		false, false)
	if err != nil || owner == nil {
		return nil, buildPreviewDisabledComment(app.Name)
	}
	var previewSettings *entity.AppFeaturePreviewSettings
	if featureSettings != nil {
		previewSettings = featureSettings.PreviewSettings
	}
	dashboardURL := ""
	if cfg := config.Current(); cfg != nil {
		dashboardURL = cfg.BaseDashboardURL()
	}
	return previewSettings, prCommentsRefusal(owner, previewSettings, dashboardURL)
}

// prCommentsRefusal is the reply refusing a comment's command on the app the
// previews are of: previews off, or comments not allowed to run commands. ""
// when neither.
func prCommentsRefusal(owner *entity.App, previewSettings *entity.AppFeaturePreviewSettings,
	dashboardURL string) string {
	switch {
	case previewSettings == nil || !previewSettings.Enabled:
		return buildPreviewDisabledComment(owner.Name)
	case !previewSettings.AllowPRComments:
		return buildPRCommentsDisabledComment(owner.Name, featureSettingsURL(dashboardURL, owner))
	}
	return ""
}

// withheldSecretsNote is the reply's warning of the secrets the previews of the
// app go without - the app's, or a preview's parent's - or "" when there are none.
// It only warns: failing to work them out leaves it out.
func (uc *UC) withheldSecretsNote(ctx context.Context, db database.IDB, app *entity.App) string {
	owner := app
	if app.IsPreviewApp() {
		var err error
		owner, err = uc.appService.LoadApp(ctx, db, app.ProjectID, app.ParentID, false, false,
			bunex.SelectExcludeColumns(entity.AppDefaultExcludeColumns...),
			bunex.SelectRelation("Project",
				bunex.SelectExcludeColumns(entity.ProjectDefaultExcludeColumns...),
			),
			bunex.SelectRelation("ProjectEnv"),
		)
		if err != nil {
			logging.Warnf("webhook: loading the parent of preview %s: %v", app.ID, err)
			return ""
		}
	}
	secrets, err := uc.appPreviewService.WithheldSecrets(ctx, db, owner)
	if err != nil {
		logging.Warnf("webhook: working out the secrets a preview of app %s goes without: %v", owner.ID, err)
		return ""
	}
	dashboardURL := ""
	if cfg := config.Current(); cfg != nil {
		dashboardURL = cfg.BaseDashboardURL()
	}
	return buildWithheldSecretsNote(owner.Name, appPageURL(dashboardURL, owner, "secrets"), secrets)
}

// featureSettingsURL is the dashboard's page of an app's feature settings, or
// "" when the dashboard's address is not known.
func featureSettingsURL(dashboardURL string, app *entity.App) string {
	return appPageURL(dashboardURL, app, "feature-settings")
}

// appPageURL is a page of an app on the dashboard, such as its secrets, or ""
// when the dashboard's address is not known.
func appPageURL(dashboardURL string, app *entity.App, page string) string {
	projectID, env := projecthelper.ParseProjectEnvID(app.ProjectEnvID)
	if dashboardURL == "" || projectID == "" || env == "" {
		return ""
	}
	u, err := url.JoinPath(dashboardURL, "projects", projectID, env, "apps", app.ID, page)
	if err != nil {
		return ""
	}
	return u + "/"
}

//nolint:gocognit,gocyclo
func (uc *UC) parsePRCommentCommand(
	commentEvent *repoPRCommentEventData,
) (bool, bool, string, error) {
	var firstValidLine string
	for _, line := range strings.Split(commentEvent.CommentBody, "\n") {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		firstValidLine = line
		break
	}

	if !strings.HasPrefix(firstValidLine, "/hivepaas") {
		return false, false, "", nil
	}

	rawCmd := firstValidLine
	fields := strings.Fields(firstValidLine)
	if len(fields) <= 1 {
		return true, false, rawCmd, nil
	}

	for _, field := range fields[1:] {
		k, v, _ := strings.Cut(field, "=")
		switch {
		case k == previewCmdDeploy || k == previewCmdCancel:
			commentEvent.previewCmd = k
		case (k == previewCmdDeployArgNoStart || k == "no-start") && commentEvent.previewCmd == previewCmdDeploy:
			if v == "" {
				commentEvent.previewDeployNoStart = true
				continue
			}
			boolVal, parseErr := strconv.ParseBool(v)
			if parseErr != nil {
				return true, false, rawCmd, hperrors.Wrap(parseErr)
			}
			commentEvent.previewDeployNoStart = boolVal
		case (k == previewCmdDeployArgNoWait || k == "no-wait") && commentEvent.previewCmd == previewCmdDeploy:
			if v == "" {
				commentEvent.previewDeployNoWait = true
				continue
			}
			boolVal, parseErr := strconv.ParseBool(v)
			if parseErr != nil {
				return true, false, rawCmd, hperrors.Wrap(parseErr)
			}
			commentEvent.previewDeployNoWait = boolVal
		case (k == previewCmdDeployArgCloneDb || k == "clone-db") && commentEvent.previewCmd == previewCmdDeploy:
			if v == "" {
				commentEvent.previewDeployCloneDB = true
				continue
			}
			boolVal, parseErr := strconv.ParseBool(v)
			if parseErr != nil {
				return true, false, rawCmd, hperrors.Wrap(parseErr)
			}
			commentEvent.previewDeployCloneDB = boolVal
		case (k == previewCmdDeployArgNoCloneDb || k == "no-clone-db") && commentEvent.previewCmd == previewCmdDeploy:
			if v == "" {
				commentEvent.previewDeployNoCloneDB = true
				continue
			}
			boolVal, parseErr := strconv.ParseBool(v)
			if parseErr != nil {
				return true, false, rawCmd, hperrors.Wrap(parseErr)
			}
			commentEvent.previewDeployNoCloneDB = boolVal
		case k == previewCmdDeployArgSubdomain && commentEvent.previewCmd == previewCmdDeploy:
			commentEvent.previewDeploySubdomain = v
		default:
			return true, false, rawCmd, nil
		}
	}

	if commentEvent.previewCmd == "" {
		return true, false, rawCmd, nil
	}

	return true, true, rawCmd, nil
}
