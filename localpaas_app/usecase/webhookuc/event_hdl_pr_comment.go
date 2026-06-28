package webhookuc

import (
	"context"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/gitsight/go-vcsurl"

	"github.com/localpaas/localpaas/localpaas_app/apperrors"
	"github.com/localpaas/localpaas/localpaas_app/base"
	"github.com/localpaas/localpaas/localpaas_app/entity"
	"github.com/localpaas/localpaas/localpaas_app/infra/database"
	"github.com/localpaas/localpaas/localpaas_app/pkg/bunex"
	"github.com/localpaas/localpaas/localpaas_app/pkg/githelper"
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
	db database.IDB,
	prCommentEvent *repoPRCommentEventData,
	data *handleRepoWebhookData,
) (err error) {
	parsedURL, err := vcsurl.Parse(prCommentEvent.RepoURL)
	if err != nil {
		return apperrors.New(err)
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
		return apperrors.New(err)
	}

	var wg sync.WaitGroup
	for _, app := range apps {
		wg.Go(func() {
			switch prCommentEvent.previewCmd {
			case previewCmdDeploy:
				if app.ParentID == "" {
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
