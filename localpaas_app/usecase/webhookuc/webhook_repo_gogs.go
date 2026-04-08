package webhookuc

import (
	"errors"

	"github.com/go-playground/webhooks/v6/gogs"
	client "github.com/gogits/go-gogs-client"

	"github.com/localpaas/localpaas/localpaas_app/apperrors"
	"github.com/localpaas/localpaas/localpaas_app/usecase/webhookuc/webhookdto"
)

func (uc *UC) processGogsWebhook(
	req *webhookdto.HandleRepoWebhookReq,
	data *repoEventData,
) error {
	hook, err := gogs.New()
	if err != nil {
		return apperrors.Wrap(err)
	}
	payload, err := hook.Parse(req.Request, gogs.PushEvent)
	if err != nil {
		if errors.Is(err, gogs.ErrEventNotFound) { // ok event wasn't one of the ones asked to be parsed
			return nil
		}
		return apperrors.Wrap(err)
	}

	switch payload.(type) { //nolint
	case client.PushPayload:
		push, _ := payload.(client.PushPayload) //nolint
		data.Push = &repoPushEventData{
			RepoRef:  push.Ref,
			RepoURL:  push.Repo.HTMLURL,
			ChangeID: push.After,
		}
	}
	return nil
}
