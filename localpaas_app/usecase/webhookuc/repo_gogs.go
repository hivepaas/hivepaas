package webhookuc

import (
	"errors"
	"net/http"

	"github.com/go-playground/webhooks/v6/gogs"
	client "github.com/gogits/go-gogs-client"

	"github.com/localpaas/localpaas/localpaas_app/apperrors"
)

func (uc *UC) parseGogsWebhook(
	req *http.Request,
	secret string,
	data *repoEventData,
) error {
	hook, err := gogs.New(gogs.Options.Secret(secret))
	if err != nil {
		return apperrors.New(err)
	}
	payload, err := hook.Parse(req, gogs.PushEvent, gogs.IssueCommentEvent, gogs.PullRequestEvent)
	if err != nil {
		if errors.Is(err, gogs.ErrEventNotFound) { // ok event wasn't one of the ones asked to be parsed
			return nil
		}
		return apperrors.New(err)
	}

	switch p := payload.(type) { //nolint
	case client.PushPayload:
		push, _ := payload.(client.PushPayload) //nolint
		data.Push = &repoPushEventData{
			RepoRef:  push.Ref,
			RepoURL:  push.Repo.HTMLURL,
			ChangeID: push.After,
		}
	case client.IssueCommentPayload:
		if string(p.Action) == actionCreated && p.Issue.PullRequest != nil {
			data.PRComment = &repoPRCommentEventData{
				RepoURL:     p.Repository.HTMLURL,
				PRNumber:    p.Issue.Index,
				CommentBody: p.Comment.Body,
			}
		}
	case client.PullRequestPayload:
		if string(p.Action) == actionClosed {
			data.PRClosed = &repoPRClosedEventData{
				RepoURL:  p.Repository.HTMLURL,
				PRNumber: p.Index,
			}
		}
	}
	return nil
}
