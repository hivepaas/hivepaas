package webhookuc

import (
	"errors"
	"net/http"
	"strings"

	"github.com/go-playground/webhooks/v6/gitea"

	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
)

func (uc *UC) parseGiteaWebhook(
	req *http.Request,
	secret string,
	data *repoEventData,
) error {
	hook, err := gitea.New(gitea.Options.Secret(secret))
	if err != nil {
		return hperrors.Wrap(err)
	}
	payload, err := hook.Parse(req, gitea.PushEvent, gitea.IssueCommentEvent,
		gitea.PullRequestCommentEvent, gitea.PullRequestEvent)
	if err != nil {
		if errors.Is(err, gitea.ErrEventNotFound) { // ok event wasn't one of the ones asked to be parsed
			return nil
		}
		return hperrors.Wrap(err)
	}

	switch p := payload.(type) { //nolint
	case gitea.PushPayload:
		push, _ := payload.(gitea.PushPayload) //nolint
		data.Push = &repoPushEventData{
			RepoRef:  push.Ref,
			RepoURL:  push.Repo.HTMLURL,
			ChangeID: push.After,
		}
	case gitea.IssueCommentPayload:
		if p.Action == actionCreated && p.IsPull {
			data.PRComment = &repoPRCommentEventData{
				RepoURL:     p.Repository.HTMLURL,
				PRNumber:    p.Issue.Index,
				CommentBody: p.Comment.Body,
				Author:      giteaCommentAuthor(&p),
			}
		}
	case gitea.PullRequestPayload:
		switch p.Action {
		case actionSynchronize:
			data.PRSynchronized = &repoPRSynchronizedEventData{
				RepoURL:  p.Repository.HTMLURL,
				PRNumber: p.Index,
				ChangeID: p.PullRequest.Head.Sha,
			}
		case actionClosed:
			data.PRClosed = &repoPRClosedEventData{
				RepoURL:  p.Repository.HTMLURL,
				PRNumber: p.Index,
			}
		}
	}
	return nil
}

// giteaCommentAuthor is who wrote a comment, as Gitea's webhook says. Whether
// they may write is asked of Gitea: its webhook says only who owns the
// repository.
func giteaCommentAuthor(p *gitea.IssueCommentPayload) prCommentAuthor {
	var author prCommentAuthor
	if p.Comment != nil && p.Comment.Poster != nil {
		author.Login = p.Comment.Poster.UserName
	}
	if p.Repository != nil {
		author.RepoPrivate = p.Repository.Private
		author.IsRepoOwner = author.Login != "" && p.Repository.Owner != nil &&
			strings.EqualFold(author.Login, p.Repository.Owner.UserName)
	}
	return author
}
