package webhookuc

import (
	"errors"
	"net/http"

	"github.com/go-playground/webhooks/v6/bitbucket"

	"github.com/localpaas/localpaas/localpaas_app/apperrors"
	"github.com/localpaas/localpaas/localpaas_app/pkg/githelper"
)

func (uc *UC) parseBitbucketWebhook(
	req *http.Request,
	secret string,
	data *repoEventData,
) error {
	hook, err := bitbucket.New(bitbucket.Options.UUID(secret))
	if err != nil {
		return apperrors.New(err)
	}
	payload, err := hook.Parse(req, bitbucket.RepoPushEvent, bitbucket.PullRequestCommentCreatedEvent,
		bitbucket.PullRequestMergedEvent, bitbucket.PullRequestDeclinedEvent)
	if err != nil {
		if errors.Is(err, bitbucket.ErrEventNotFound) { // ok event wasn't one of the ones asked to be parsed
			return nil
		}
		return apperrors.New(err)
	}

	switch p := payload.(type) { //nolint
	case bitbucket.RepoPushPayload:
		push, _ := payload.(bitbucket.RepoPushPayload) //nolint
		lastChange := push.Push.Changes[len(push.Push.Changes)-1]
		data.Push = &repoPushEventData{
			RepoRef:  string(githelper.NormalizeRepoRef(lastChange.New.Name)),
			RepoURL:  push.Repository.Links.HTML.Href,
			ChangeID: lastChange.New.Target.Hash,
		}
	case bitbucket.PullRequestCommentCreatedPayload:
		data.PRComment = &repoPRCommentEventData{
			RepoURL:     p.Repository.Links.HTML.Href,
			PRNumber:    p.PullRequest.ID,
			CommentBody: p.Comment.Content.Raw,
			Branch:      "heads/" + p.PullRequest.Source.Branch.Name,
		}
	case bitbucket.PullRequestMergedPayload:
		data.PRClosed = &repoPRClosedEventData{
			RepoURL:  p.Repository.Links.HTML.Href,
			PRNumber: p.PullRequest.ID,
			Branch:   "heads/" + p.PullRequest.Source.Branch.Name,
		}
	case bitbucket.PullRequestDeclinedPayload:
		data.PRClosed = &repoPRClosedEventData{
			RepoURL:  p.Repository.Links.HTML.Href,
			PRNumber: p.PullRequest.ID,
			Branch:   "heads/" + p.PullRequest.Source.Branch.Name,
		}
	}
	return nil
}
