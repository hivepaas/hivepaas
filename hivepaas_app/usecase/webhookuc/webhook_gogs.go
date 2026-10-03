package webhookuc

import (
	"errors"
	"net/http"
	"strings"

	"github.com/go-playground/webhooks/v6/gogs"
	client "github.com/gogits/go-gogs-client"
	"github.com/tiendc/gofn"

	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
)

func (uc *UC) parseGogsWebhook(
	req *http.Request,
	secret string,
	data *repoEventData,
) error {
	hook, err := gogs.New(gogs.Options.Secret(secret))
	if err != nil {
		return hperrors.Wrap(err)
	}
	payload, err := hook.Parse(req, gogs.PushEvent, gogs.IssueCommentEvent, gogs.PullRequestEvent)
	if err != nil {
		if errors.Is(err, gogs.ErrEventNotFound) { // ok event wasn't one of the ones asked to be parsed
			return nil
		}
		return hperrors.Wrap(err)
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
				Author:      gogsCommentAuthor(&p),
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

// gogsCommentAuthor is who wrote a comment, as Gogs' webhook says: HivePaaS
// cannot ask Gogs whether they may write, so only the repository's owner, or
// anyone on a private repository, may run commands.
func gogsCommentAuthor(p *client.IssueCommentPayload) prAuthor {
	var author prAuthor
	if p.Sender != nil {
		author.Login = gofn.Coalesce(p.Sender.UserName, p.Sender.Login)
	}
	if p.Repository != nil {
		author.RepoPrivate = p.Repository.Private
		if owner := p.Repository.Owner; owner != nil {
			author.IsRepoOwner = author.Login != "" &&
				strings.EqualFold(author.Login, gofn.Coalesce(owner.UserName, owner.Login))
		}
	}
	return author
}
