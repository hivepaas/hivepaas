package webhookuc

import (
	"context"
	"sync"

	"github.com/gitsight/go-vcsurl"

	"github.com/localpaas/localpaas/localpaas_app/apperrors"
	"github.com/localpaas/localpaas/localpaas_app/entity"
	"github.com/localpaas/localpaas/localpaas_app/infra/database"
	"github.com/localpaas/localpaas/localpaas_app/pkg/bunex"
)

type repoPushEventData struct {
	RepoRef  string
	RepoURL  string
	ChangeID string
}

func (uc *UC) processWebhookEventPush(
	ctx context.Context,
	db database.IDB,
	pushEvent *repoPushEventData,
	data *handleRepoWebhookData,
) (err error) {
	parsedURL, err := vcsurl.Parse(pushEvent.RepoURL)
	if err != nil {
		return apperrors.New(err)
	}

	apps, err := uc.appService.FindAppsMatchingRepository(ctx, db, parsedURL.ID, pushEvent.RepoRef,
		bunex.SelectExcludeColumns(entity.AppDefaultExcludeColumns...),
	)
	if err != nil {
		return apperrors.New(err)
	}
	var wg sync.WaitGroup
	for _, app := range apps {
		wg.Go(func() {
			_ = uc.createAppDeployment(ctx, app, pushEvent.ChangeID, data.WebhookSetting.ID)
		})
	}
	wg.Wait()
	return nil
}
