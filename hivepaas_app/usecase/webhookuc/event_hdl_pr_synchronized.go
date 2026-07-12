package webhookuc

import (
	"context"
	"sync"

	"github.com/gitsight/go-vcsurl"

	"github.com/hivepaas/hivepaas/hivepaas_app/apperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/infra/database"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/bunex"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/githelper"
)

type repoPRSynchronizedEventData struct {
	RepoURL  string
	PRNumber int64
	ChangeID string
}

// processWebhookEventPRSynchronized handles pull request synchronization/update events
func (uc *UC) processWebhookEventPRSynchronized(
	ctx context.Context,
	db database.IDB,
	event *repoPRSynchronizedEventData,
	data *handleRepoWebhookData,
) error {
	parsedURL, err := vcsurl.Parse(event.RepoURL)
	if err != nil {
		return apperrors.Wrap(err)
	}

	webhook := data.WebhookSetting.MustAsRepoWebhook()
	expectedRef, _ := githelper.GetPullNumberRef(event.PRNumber, base.GitSource(webhook.Kind))
	if expectedRef == "" {
		return nil
	}

	// We look for preview apps (which have parent_id IS NOT NULL) matching the repository
	apps, err := uc.appService.FindAppsMatchingRepository(ctx, db, parsedURL.ID, expectedRef,
		bunex.SelectWhere("app.parent_id IS NOT NULL"),
		bunex.SelectExcludeColumns(entity.AppDefaultExcludeColumns...),
	)
	if err != nil {
		return apperrors.Wrap(err)
	}

	var wg sync.WaitGroup
	for _, app := range apps {
		if app.ParentID == "" { // The app is not a preview, skip it. Just recheck for safety.
			continue
		}
		wg.Go(func() {
			_ = uc.createAppDeployment(ctx, app, event.ChangeID, data.WebhookSetting.ID)
		})
	}
	wg.Wait()

	return nil
}
