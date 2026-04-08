package webhookuc

import (
	"errors"

	"github.com/go-playground/webhooks/v6/bitbucket"

	"github.com/localpaas/localpaas/localpaas_app/apperrors"
	"github.com/localpaas/localpaas/localpaas_app/pkg/githelper"
	"github.com/localpaas/localpaas/localpaas_app/usecase/webhookuc/webhookdto"
)

func (uc *UC) processBitbucketWebhook(
	req *webhookdto.HandleRepoWebhookReq,
	data *repoEventData,
) error {
	hook, err := bitbucket.New()
	if err != nil {
		return apperrors.Wrap(err)
	}
	payload, err := hook.Parse(req.Request, bitbucket.RepoPushEvent)
	if err != nil {
		if errors.Is(err, bitbucket.ErrEventNotFound) { // ok event wasn't one of the ones asked to be parsed
			return nil
		}
		return apperrors.Wrap(err)
	}

	switch payload.(type) { //nolint
	case bitbucket.RepoPushPayload:
		push, _ := payload.(bitbucket.RepoPushPayload) //nolint
		data.Push = &repoPushEventData{
			RepoRef:  string(githelper.NormalizeRepoRef(push.Push.Changes[0].New.Name)),
			RepoURL:  push.Repository.Links.HTML.Href,
			ChangeID: push.Push.Changes[len(push.Push.Changes)-1].New.Target.Hash,
		}
	}
	return nil
}
