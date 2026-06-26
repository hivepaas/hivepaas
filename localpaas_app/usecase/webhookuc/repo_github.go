package webhookuc

import (
	"errors"
	"net/http"

	"github.com/go-playground/webhooks/v6/github"

	"github.com/localpaas/localpaas/localpaas_app/apperrors"
)

const (
	actionCreated = "created"
	actionClosed  = "closed"
)

func (uc *UC) parseGithubWebhook(
	req *http.Request,
	secret string,
	data *repoEventData,
) error {
	hook, err := github.New(github.Options.Secret(secret))
	if err != nil {
		return apperrors.New(err)
	}
	payload, err := hook.Parse(req, github.PushEvent, github.IssueCommentEvent, github.PullRequestEvent)
	if err != nil {
		if errors.Is(err, github.ErrEventNotFound) { // ok event wasn't one of the ones asked to be parsed
			return nil
		}
		return apperrors.New(err)
	}

	switch p := payload.(type) { //nolint
	case github.PushPayload:
		push, _ := payload.(github.PushPayload) //nolint
		data.Push = &repoPushEventData{
			RepoRef:  push.Ref,
			RepoURL:  push.Repository.HTMLURL,
			ChangeID: push.After,
		}
	case github.IssueCommentPayload:
		if p.Action == actionCreated && p.Issue.PullRequest != nil {
			data.PRComment = &repoPRCommentEventData{
				RepoURL:     p.Repository.HTMLURL,
				PRNumber:    p.Issue.Number,
				CommentBody: p.Comment.Body,
			}
		}
	case github.PullRequestPayload:
		if p.Action == actionClosed {
			data.PRClosed = &repoPRClosedEventData{
				RepoURL:  p.Repository.HTMLURL,
				PRNumber: p.Number,
			}
		}
	}
	return nil
}
