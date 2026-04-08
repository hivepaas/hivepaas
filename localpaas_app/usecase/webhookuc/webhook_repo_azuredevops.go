package webhookuc

import (
	"strings"

	"github.com/go-playground/webhooks/v6/azuredevops"

	"github.com/localpaas/localpaas/localpaas_app/apperrors"
	"github.com/localpaas/localpaas/localpaas_app/usecase/webhookuc/webhookdto"
)

func (uc *UC) processAzureDevOpsWebhook(
	req *webhookdto.HandleRepoWebhookReq,
	data *repoEventData,
) error {
	hook, err := azuredevops.New()
	if err != nil {
		return apperrors.Wrap(err)
	}
	payload, err := hook.Parse(req.Request, azuredevops.GitPushEventType)
	if err != nil {
		if strings.Contains(err.Error(), "unknown event ") {
			return nil
		}
		return apperrors.Wrap(err)
	}

	switch payload.(type) { //nolint
	case azuredevops.GitPushEvent:
		push, _ := payload.(azuredevops.GitPushEvent) //nolint
		refUpdate := push.Resource.RefUpdates[len(push.Resource.RefUpdates)-1]
		data.Push = &repoPushEventData{
			RepoRef:  refUpdate.Name,
			RepoURL:  push.Resource.Repository.RemoteURL,
			ChangeID: refUpdate.NewObjectID,
		}
	}
	return nil
}
