package webhookuc

import (
	"errors"

	"github.com/go-playground/webhooks/v6/gitlab"

	"github.com/localpaas/localpaas/localpaas_app/apperrors"
	"github.com/localpaas/localpaas/localpaas_app/usecase/webhookuc/webhookdto"
)

func (uc *UC) processGitlabWebhook(
	req *webhookdto.HandleRepoWebhookReq,
	data *repoEventData,
) error {
	hook, err := gitlab.New()
	if err != nil {
		return apperrors.Wrap(err)
	}
	payload, err := hook.Parse(req.Request, gitlab.PushEvents)
	if err != nil {
		if errors.Is(err, gitlab.ErrEventNotFound) { // ok event wasn't one of the ones asked to be parsed
			return nil
		}
		return apperrors.Wrap(err)
	}

	switch payload.(type) { //nolint
	case gitlab.PushEventPayload:
		push, _ := payload.(gitlab.PushEventPayload) //nolint
		data.Push = &repoPushEventData{
			RepoRef:  push.Ref,
			RepoURL:  push.Repository.GitHTTPURL,
			ChangeID: push.After,
		}
	}
	return nil
}
