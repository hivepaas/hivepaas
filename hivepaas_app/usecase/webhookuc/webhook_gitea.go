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
		// Gitea says synchronized; synchronize is kept for any that say it as
		// GitHub does.
		case actionSynchronize, giteaActionSynchronized:
			data.PRSynchronized = &repoPRSynchronizedEventData{
				RepoURL:  p.Repository.HTMLURL,
				PRNumber: p.Index,
				ChangeID: p.PullRequest.Head.Sha,
				Author:   giteaAuthor(p.PullRequest.Poster, p.Repository),
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

// giteaActionSynchronized is the action of Gitea's pull request event when
// commits are pushed to it.
const giteaActionSynchronized = "synchronized"

// giteaCommentAuthor is who wrote a comment, as Gitea's webhook says. Whether
// they may write is asked of Gitea: its webhook says only who owns the
// repository.
func giteaCommentAuthor(p *gitea.IssueCommentPayload) prAuthor {
	var poster *gitea.User
	if p.Comment != nil {
		poster = p.Comment.Poster
	}
	return giteaAuthor(poster, p.Repository)
}

// giteaAuthor is a user of a repository - a comment's or a pull request's
// author - as Gitea's webhook says.
func giteaAuthor(user *gitea.User, repo *gitea.Repository) prAuthor {
	var author prAuthor
	if user != nil {
		author.Login = user.UserName
	}
	if repo != nil {
		author.RepoPrivate = repo.Private
		author.IsRepoOwner = author.Login != "" && repo.Owner != nil &&
			strings.EqualFold(author.Login, repo.Owner.UserName)
	}
	return author
}
