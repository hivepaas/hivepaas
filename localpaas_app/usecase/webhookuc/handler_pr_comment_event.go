package webhookuc

import (
	"context"
	"strconv"
	"strings"

	"github.com/localpaas/localpaas/localpaas_app/apperrors"
	"github.com/localpaas/localpaas/localpaas_app/base"
	"github.com/localpaas/localpaas/localpaas_app/entity"
	"github.com/localpaas/localpaas/localpaas_app/infra/database"
	"github.com/localpaas/localpaas/localpaas_app/service/apppreviewservice"
	"github.com/localpaas/localpaas/localpaas_app/service/appservice"
)

type repoPRCommentEventData struct {
	RepoID      string
	RepoURL     string
	PRNumber    int64
	CommentBody string
	Branch      string

	// Parsed command data
	previewDeploy    string
	previewNoStart   bool
	previewSubdomain string
}

func (uc *UC) createAppPreviewByCommentEvent(
	ctx context.Context,
	db database.Tx,
	app *entity.App,
	commentEvent *repoPRCommentEventData,
	data *handleRepoWebhookData,
	persistingData *appservice.PersistingAppData,
) error {
	var repoRef string
	prNumberStr := strconv.FormatInt(commentEvent.PRNumber, 10)
	if commentEvent.Branch != "" {
		repoRef = commentEvent.Branch
	} else {
		webhook := data.WebhookSetting.MustAsRepoWebhook()
		switch webhook.Kind {
		case base.WebhookKindGithub:
			repoRef = "pull/" + prNumberStr + "/head"
		case base.WebhookKindGitlab:
			repoRef = "merge-requests/" + prNumberStr + "/head"
		case base.WebhookKindGitea, base.WebhookKindGogs:
			repoRef = "pull/" + prNumberStr + "/head"
		case base.WebhookKindBitbucket:
			fallthrough
		default:
			return apperrors.New(apperrors.ErrWebhookTypeUnsupported).WithParam("Type", webhook.Kind)
		}
	}

	createResp, err := uc.appPreviewService.CreatePreview(ctx, db, &apppreviewservice.CreatePreviewReq{
		ProjectID:       app.ProjectID,
		AppID:           app.ID,
		RepoRef:         repoRef,
		NoStart:         commentEvent.previewNoStart,
		CustomSubdomain: commentEvent.previewSubdomain,
		OnInitDeployment: func(deployment *entity.Deployment) error {
			deployment.Trigger = &entity.AppDeploymentTrigger{
				Source:   base.DeploymentTriggerSourceRepoWebhook,
				SourceID: data.WebhookSetting.ID,
				ChangeID: "pr-" + prNumberStr,
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

	for _, field := range fields {
		k, v, _ := strings.Cut(field, "=")
		switch k {
		case "deploy":
			commentEvent.previewDeploy = field
		case "nostart", "no-start":
			boolVal, err := strconv.ParseBool(v)
			if err != nil {
				return false, apperrors.New(err)
			}
			commentEvent.previewNoStart = boolVal
		case "subdomain":
			commentEvent.previewSubdomain = v
		}
	}

	return commentEvent.previewDeploy != "", nil
}
