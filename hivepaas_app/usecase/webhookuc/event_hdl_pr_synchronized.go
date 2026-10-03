package webhookuc

import (
	"context"
	"sync"

	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/infra/database"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/bunex"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/githelper"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/logging"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/safego"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/vcsurl"
)

type repoPRSynchronizedEventData struct {
	RepoURL  string
	PRNumber int64
	ChangeID string
	// Author is who opened the pull request: its previews follow its new
	// commits only when they may write to the repository.
	Author prAuthor
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
		return hperrors.Wrap(err)
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
		return hperrors.Wrap(err)
	}

	if len(apps) == 0 {
		return nil
	}

	// A preview runs the pull request's code with the app's env vars: a
	// stranger's new commits wait for someone who may write to deploy them.
	if !uc.prAuthorAllowed(ctx, db, event.RepoURL, &event.Author, data, apps) {
		logging.Warnf("webhook: %s may not write to %s: the new commits of pull request %d are not deployed",
			event.Author.Login, event.RepoURL, event.PRNumber)
		_ = uc.sendPRComment(ctx, db, &repoPRCommentEventData{RepoURL: event.RepoURL, PRNumber: event.PRNumber},
			data, apps[0], buildPushNotDeployedComment(event.ChangeID))
		return nil
	}

	var wg sync.WaitGroup
	for _, app := range apps {
		if !app.IsPreviewApp() { // The app is not a preview, skip it. Just recheck for safety.
			continue
		}
		wg.Go(func() {
			defer safego.Recover("webhook.prSynchronized.createAppDeployment")
			_ = uc.createAppDeployment(ctx, app, event.ChangeID, data.WebhookSetting.ID)
		})
	}
	wg.Wait()

	return nil
}
