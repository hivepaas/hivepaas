package webhookuc

import (
	"context"
	"strconv"
	"strings"
	"sync"

	"github.com/gitsight/go-vcsurl"

	"github.com/hivepaas/hivepaas/hivepaas_app/apperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/infra/database"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/bunex"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/githelper"
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
		return apperrors.Wrap(err)
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
		return apperrors.Wrap(err)
	}

	var firstApp *entity.App
	if len(apps) > 0 {
		firstApp = apps[0]
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
	if !app.IsChildApp() {
		previewSettings, err := uc.loadAppPreviewSettings(ctx, db, app)
		if err != nil || previewSettings == nil || !previewSettings.Enabled {
			_ = uc.sendPRComment(ctx, db, prCommentEvent, data, app, buildPreviewDisabledComment(app.Name))
			return
		}

		err = uc.createAppPreview(ctx, app, prCommentEvent, repoRef, data.WebhookSetting.ID, previewSettings)
		if err != nil {
			_ = uc.sendPRComment(ctx, db, prCommentEvent, data, app, buildDeployFailedComment(app.Name, err))
			return
		}

		cloneDBApps := prCommentEvent.previewDeployCloneDB
		if !prCommentEvent.previewDeployNoCloneDB {
			cloneDBApps = cloneDBApps || (previewSettings.AutoCloneApps && len(previewSettings.AppsToClone) > 0)
		}
		_ = uc.sendPRComment(ctx, db, prCommentEvent, data, app, buildDeployPreviewComment(cloneDBApps))
	} else {
		// TODO: find the SHA of the head commit of the PR (change id)
		err := uc.createAppDeployment(ctx, app, "", data.WebhookSetting.ID)
		if err != nil {
			_ = uc.sendPRComment(ctx, db, prCommentEvent, data, app, buildDeployFailedComment(app.Name, err))
			return
		}
		_ = uc.sendPRComment(ctx, db, prCommentEvent, data, app, buildDeployPreviewComment(true))
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
	err := uc.deleteAppPreview(ctx, app, repoRef)
	if err != nil {
		_ = uc.sendPRComment(ctx, db, prCommentEvent, data, app, buildDeployFailedComment(app.Name, err))
		return
	}
	_ = uc.sendPRComment(ctx, db, prCommentEvent, data, app, buildCancelPreviewComment())
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
				return true, false, rawCmd, apperrors.Wrap(parseErr)
			}
			commentEvent.previewDeployNoStart = boolVal
		case (k == previewCmdDeployArgNoWait || k == "no-wait") && commentEvent.previewCmd == previewCmdDeploy:
			if v == "" {
				commentEvent.previewDeployNoWait = true
				continue
			}
			boolVal, parseErr := strconv.ParseBool(v)
			if parseErr != nil {
				return true, false, rawCmd, apperrors.Wrap(parseErr)
			}
			commentEvent.previewDeployNoWait = boolVal
		case (k == previewCmdDeployArgCloneDb || k == "clone-db") && commentEvent.previewCmd == previewCmdDeploy:
			if v == "" {
				commentEvent.previewDeployCloneDB = true
				continue
			}
			boolVal, parseErr := strconv.ParseBool(v)
			if parseErr != nil {
				return true, false, rawCmd, apperrors.Wrap(parseErr)
			}
			commentEvent.previewDeployCloneDB = boolVal
		case (k == previewCmdDeployArgNoCloneDb || k == "no-clone-db") && commentEvent.previewCmd == previewCmdDeploy:
			if v == "" {
				commentEvent.previewDeployNoCloneDB = true
				continue
			}
			boolVal, parseErr := strconv.ParseBool(v)
			if parseErr != nil {
				return true, false, rawCmd, apperrors.Wrap(parseErr)
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
