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

type repoPRClosedEventData struct {
	RepoURL  string
	PRNumber int64
	Branch   string // populated for Bitbucket
}

func (uc *UC) processWebhookEventPRClosed(
	ctx context.Context,
	db database.IDB,
	prClosedEvent *repoPRClosedEventData,
	data *handleRepoWebhookData,
) (err error) {
	parsedURL, err := vcsurl.Parse(prClosedEvent.RepoURL)
	if err != nil {
		return apperrors.New(err)
	}

	var expectedRef string
	webhook := data.WebhookSetting.MustAsRepoWebhook()
	if webhook.Kind == base.WebhookKindBitbucket && prClosedEvent.Branch != "" {
		expectedRef = string(githelper.NormalizeRepoRef(prClosedEvent.Branch))
	}
	if expectedRef == "" {
		expectedRef, _ = githelper.GetPullNumberRef(prClosedEvent.PRNumber, base.GitSource(webhook.Kind))
	}
	if expectedRef == "" {
		return nil
	}

	apps, err := uc.appService.FindAppsMatchingRepository(ctx, db, parsedURL.ID, expectedRef,
		bunex.SelectWhere("app.parent_id IS NOT NULL"), // Load preview apps to delete
		bunex.SelectExcludeColumns(entity.AppDefaultExcludeColumns...),
	)
	if err != nil {
		return apperrors.New(err)
	}

	var wg sync.WaitGroup
	for _, app := range apps {
		wg.Go(func() {
			_ = uc.deleteAppPreview(ctx, app, expectedRef)
		})
	}
	wg.Wait()

	return nil
}
