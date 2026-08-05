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

	success, _ := uc.parsePRCommentCommand(prCommentEvent)
	if !success {
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

	var wg sync.WaitGroup
	for _, app := range apps {
		wg.Go(func() {
			switch prCommentEvent.previewCmd {
			case previewCmdDeploy:
				if !app.IsChildApp() {
					_ = uc.createAppPreview(ctx, app, prCommentEvent, repoRef, data.WebhookSetting.ID)
				} else {
					// TODO: find the SHA of the head commit of the PR (change id)
					_ = uc.createAppDeployment(ctx, app, "", data.WebhookSetting.ID)
				}
			case previewCmdCancel:
				_ = uc.deleteAppPreview(ctx, app, repoRef)
			}
		})
	}
	wg.Wait()

	return nil
}

//nolint:gocognit,gocyclo
func (uc *UC) parsePRCommentCommand(
	commentEvent *repoPRCommentEventData,
) (success bool, err error) {
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
				return false, apperrors.Wrap(err)
			}
			commentEvent.previewDeployNoStart = boolVal
		case (k == previewCmdDeployArgNoWait || k == "no-wait") && commentEvent.previewCmd == previewCmdDeploy:
			if v == "" {
				commentEvent.previewDeployNoWait = true
				continue // continue for-loop
			}
			boolVal, err := strconv.ParseBool(v)
			if err != nil {
				return false, apperrors.Wrap(err)
			}
			commentEvent.previewDeployNoWait = boolVal
		case (k == previewCmdDeployArgCloneDb || k == "clone-db") && commentEvent.previewCmd == previewCmdDeploy:
			if v == "" {
				commentEvent.previewDeployCloneDB = true
				continue // continue for-loop
			}
			boolVal, err := strconv.ParseBool(v)
			if err != nil {
				return false, apperrors.Wrap(err)
			}
			commentEvent.previewDeployCloneDB = boolVal
		case (k == previewCmdDeployArgNoCloneDb || k == "no-clone-db") && commentEvent.previewCmd == previewCmdDeploy:
			if v == "" {
				commentEvent.previewDeployNoCloneDB = true
				continue // continue for-loop
			}
			boolVal, err := strconv.ParseBool(v)
			if err != nil {
				return false, apperrors.Wrap(err)
			}
			commentEvent.previewDeployNoCloneDB = boolVal
		case k == previewCmdDeployArgSubdomain && commentEvent.previewCmd == previewCmdDeploy:
			commentEvent.previewDeploySubdomain = v
		}
	}

	success = commentEvent.previewCmd != ""
	if !success {
		return false, nil
	}

	// TODO (med): send a comment to the PR to notify user about the error

	return true, nil
}
