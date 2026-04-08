package webhookuc

import (
	"errors"

	"github.com/go-playground/webhooks/v6/github"

	"github.com/localpaas/localpaas/localpaas_app/apperrors"
	"github.com/localpaas/localpaas/localpaas_app/usecase/webhookuc/webhookdto"
)

func (uc *UC) processGithubWebhook(
	req *webhookdto.HandleRepoWebhookReq,
	data *repoEventData,
) error {
	hook, err := github.New()
	if err != nil {
		return apperrors.Wrap(err)
	}
	payload, err := hook.Parse(req.Request, github.PushEvent)
	if err != nil {
		if errors.Is(err, github.ErrEventNotFound) { // ok event wasn't one of the ones asked to be parsed
			return nil
		}
		return apperrors.Wrap(err)
	}

	switch payload.(type) { //nolint
	case github.PushPayload:
		push, _ := payload.(github.PushPayload) //nolint
		data.Push = &repoPushEventData{
			RepoRef:  push.Ref,
			RepoURL:  push.Repository.HTMLURL,
			ChangeID: push.After,
		}
	}
	return nil
}
