package githubappuc

import (
	"context"

	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/bunex"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/logging"
	"github.com/hivepaas/hivepaas/services/git/github"
)

// ensureAppSlug fills in the slug and owner type of a GitHub App setting that
// was created before those were persisted.
//
// The slug cannot be derived from the app name (GitHub normalizes it lossily and
// appends a counter on collisions), so the only source is the API. This reads it
// once via `GET /app` and stores it, so later reads are free.
//
// It is best-effort: on any failure the setting is returned as-is with empty
// github.com URLs rather than failing the caller.
func (uc *UC) ensureAppSlug(ctx context.Context, setting *entity.Setting) {
	githubApp, err := setting.AsGithubApp()
	if err != nil || githubApp.Slug != "" {
		return
	}
	// An inherited setting is owned by another scope: leave it to its owner.
	if setting.ObjectID != setting.CurrentObjectID {
		return
	}
	if githubApp.AppID == 0 {
		return
	}

	client, err := github.NewFromSetting(setting)
	if err != nil {
		logging.Warnf("github app %s: cannot build client to read the slug: %v", setting.ID, err)
		return
	}
	app, err := client.GetApp(ctx)
	if err != nil {
		logging.Warnf("github app %s: failed to read the app slug: %v", setting.ID, err)
		return
	}
	if app.GetSlug() == "" {
		return
	}

	githubApp.Slug = app.GetSlug()
	githubApp.OwnerLogin = app.GetOwner().GetLogin()
	githubApp.OwnerType = app.GetOwner().GetType()
	setting.MustSetData(githubApp)

	// Backfill only the data column and leave UpdateVer alone: this is metadata
	// repair, not a user edit, and it must not collide with a concurrent one.
	if err := uc.SettingRepo.Update(ctx, uc.DB, setting, bunex.UpdateColumns("data")); err != nil {
		logging.Warnf("github app %s: failed to persist the app slug: %v", setting.ID, err)
	}
}
