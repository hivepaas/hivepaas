package webhookuc

import (
	"context"
	"strconv"
	"time"

	"github.com/gitsight/go-vcsurl"

	"github.com/localpaas/localpaas/localpaas_app/apperrors"
	"github.com/localpaas/localpaas/localpaas_app/base"
	"github.com/localpaas/localpaas/localpaas_app/entity"
	"github.com/localpaas/localpaas/localpaas_app/infra/database"
	"github.com/localpaas/localpaas/localpaas_app/pkg/bunex"
	"github.com/localpaas/localpaas/localpaas_app/pkg/githelper"
	"github.com/localpaas/localpaas/localpaas_app/service/appservice"
)

type repoPRClosedEventData struct {
	RepoID   string
	RepoURL  string
	PRNumber int64
	Branch   string // populated for Bitbucket
}

func (uc *UC) processWebhookEventPRClosed(
	ctx context.Context,
	db database.Tx,
	prClosedEvent *repoPRClosedEventData,
	data *handleRepoWebhookData,
	persistingData *appservice.PersistingAppData,
) (err error) {
	parsedURL, err := vcsurl.Parse(prClosedEvent.RepoURL)
	if err != nil {
		return apperrors.New(err)
	}
	prClosedEvent.RepoID = parsedURL.ID

	// Load preview apps to delete
	appListOpts := []bunex.SelectQueryOption{
		bunex.SelectWhere("app.parent_id IS NOT NULL"),
	}

	apps, err := uc.findAppsMatchingRepository(ctx, db, prClosedEvent.RepoID, "", appListOpts...)
	if err != nil {
		return apperrors.New(err)
	}
	for _, app := range apps {
		err = uc.deleteAppPreviewByPRClosedEvent(ctx, db, app, prClosedEvent, data, persistingData)
		if err != nil {
			return apperrors.New(err)
		}
	}
	return nil
}

func (uc *UC) deleteAppPreviewByPRClosedEvent(
	ctx context.Context,
	db database.Tx,
	app *entity.App,
	closedEvent *repoPRClosedEventData,
	data *handleRepoWebhookData,
	persistingData *appservice.PersistingAppData,
) error {
	shouldDelete, err := uc.shouldDeleteAppPreviewByPRClosedEvent(ctx, db, app, closedEvent)
	if err != nil {
		return apperrors.New(err)
	}
	if !shouldDelete {
		return nil
	}

	webhook := data.WebhookSetting.MustAsRepoWebhook()
	var expectedRef string
	prNumberStr := strconv.FormatInt(closedEvent.PRNumber, 10)

	switch webhook.Kind {
	case base.WebhookKindGithub:
		expectedRef = "refs/pull/" + prNumberStr + "/head"
	case base.WebhookKindGitlab:
		expectedRef = "refs/merge-requests/" + prNumberStr + "/head"
	case base.WebhookKindGitea, base.WebhookKindGogs:
		expectedRef = "refs/pull/" + prNumberStr + "/head"
	case base.WebhookKindBitbucket:
		if closedEvent.Branch != "" {
			expectedRef = string(githelper.NormalizeRepoRef(closedEvent.Branch))
		}
	}

	if expectedRef == "" {
		return nil
	}

	deploymentSetting := app.GetSettingByType(base.SettingTypeAppDeployment)
	if deploymentSetting == nil {
		return nil
	}
	deploymentSettings, err := deploymentSetting.AsAppDeploymentSettings()
	if err != nil {
		return apperrors.New(err)
	}

	if deploymentSettings.RepoSource != nil && deploymentSettings.RepoSource.RepoRef == expectedRef {
		if err = uc.appService.DeleteApp(ctx, db, app); err != nil {
			return apperrors.New(err)
		}
		app.DeletedAt = time.Now()
		app.UpdateVer++
		persistingData.UpsertingApps = append(persistingData.UpsertingApps, app)
	}

	return nil
}

//nolint:unparam
func (uc *UC) shouldDeleteAppPreviewByPRClosedEvent(
	_ context.Context,
	_ database.Tx,
	app *entity.App,
	_ *repoPRClosedEventData,
) (bool, error) {
	// The app is not a preview, skip it
	if app.ParentID == "" {
		return false, nil
	}

	// TODO: Load preview settings and check them

	return true, nil
}
