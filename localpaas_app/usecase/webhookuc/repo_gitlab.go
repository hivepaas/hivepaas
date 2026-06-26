package webhookuc

import (
	"errors"
	"net/http"

	"github.com/go-playground/webhooks/v6/gitlab"

	"github.com/localpaas/localpaas/localpaas_app/apperrors"
)

func (uc *UC) parseGitlabWebhook(
	req *http.Request,
	secret string,
	data *repoEventData,
) error {
	hook, err := gitlab.New(gitlab.Options.Secret(secret))
	if err != nil {
		return apperrors.New(err)
	}
	payload, err := hook.Parse(req, gitlab.PushEvents, gitlab.CommentEvents)
	if err != nil {
		if errors.Is(err, gitlab.ErrEventNotFound) { // ok event wasn't one of the ones asked to be parsed
			return nil
		}
		return apperrors.New(err)
	}

	switch p := payload.(type) { //nolint
	case gitlab.PushEventPayload:
		push, _ := payload.(gitlab.PushEventPayload) //nolint
		data.Push = &repoPushEventData{
			RepoRef:  push.Ref,
			RepoURL:  push.Repository.GitHTTPURL,
			ChangeID: push.After,
		}
	case gitlab.CommentEventPayload:
		if p.ObjectAttributes.NotebookType == "MergeRequest" {
			data.PRComment = &repoPRCommentEventData{
				RepoURL:     p.Repository.GitHTTPURL,
				PRNumber:    p.MergeRequest.IID,
				CommentBody: p.ObjectAttributes.Note,
			}
		}
	}
	return nil
}
