package webhookuc

import (
	"bytes"
	"encoding/json"
	"errors"
	"io"
	"net/http"

	"github.com/go-playground/webhooks/v6/github"

	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
)

const (
	actionCreated     = "created"
	actionSynchronize = "synchronize"
	actionClosed      = "closed"
)

func (uc *UC) parseGithubWebhook(
	req *http.Request,
	secret string,
	data *repoEventData,
) error {
	hook, err := github.New(github.Options.Secret(secret))
	if err != nil {
		return hperrors.Wrap(err)
	}
	// Kept for what the library's payloads leave out: a pull request's
	// author_association.
	var body []byte
	if req.Body != nil {
		if body, err = io.ReadAll(req.Body); err != nil {
			return hperrors.Wrap(err)
		}
		_ = req.Body.Close()
		req.Body = io.NopCloser(bytes.NewReader(body))
	}
	payload, err := hook.Parse(req, github.PushEvent, github.IssueCommentEvent, github.PullRequestEvent)
	if err != nil {
		if errors.Is(err, github.ErrEventNotFound) { // ok event wasn't one of the ones asked to be parsed
			return nil
		}
		return hperrors.Wrap(err)
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
				Author: prAuthor{
					Login:       p.Comment.User.Login,
					Association: p.Comment.AuthorAssociation,
				},
			}
		}
	case github.PullRequestPayload:
		switch p.Action {
		case actionSynchronize:
			data.PRSynchronized = &repoPRSynchronizedEventData{
				RepoURL:  p.Repository.HTMLURL,
				PRNumber: p.Number,
				ChangeID: p.PullRequest.Head.Sha,
				Author: prAuthor{
					Login:       p.PullRequest.User.Login,
					Association: githubPRAssociation(body),
				},
			}
		case actionClosed:
			data.PRClosed = &repoPRClosedEventData{
				RepoURL:  p.Repository.HTMLURL,
				PRNumber: p.Number,
			}
		}
	}
	return nil
}

// githubPRAssociation is the author_association of a pull request event's pull
// request, which the library's payload leaves out. NONE when it cannot be
// read: its author is then a stranger.
func githubPRAssociation(body []byte) string {
	var event struct {
		PullRequest struct {
			AuthorAssociation string `json:"author_association"`
		} `json:"pull_request"`
	}
	if json.Unmarshal(body, &event) != nil || event.PullRequest.AuthorAssociation == "" {
		return "NONE"
	}
	return event.PullRequest.AuthorAssociation
}
