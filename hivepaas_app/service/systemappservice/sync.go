package systemappservice

import (
	"context"
	"slices"

	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/infra/database"
)

// SyncedApp is one system app a sync looks at: what its settings say of it,
// and what it was before.
type SyncedApp struct {
	Key, Name string
	// Wanted says the settings want the app run.
	Wanted bool
	// Before is the app as the sync found it; nil when there was none.
	Before *entity.App
	// Recreated says the sync removed it for its service being gone, for the
	// feature's apply to provision it again.
	Recreated bool
}

// RemoveIfServiceGone loads the app, and removes it when the settings want it
// and its service is gone: a deployment cannot make one, and the feature's
// apply provisions the app again. Its data stays: removeStorage is never set.
func RemoveIfServiceGone(ctx context.Context, db database.IDB, s Service, app *SyncedApp) error {
	loaded, err := s.LoadApp(ctx, db, app.Key)
	if err != nil {
		return hperrors.Wrap(err)
	}
	app.Before = loaded
	if loaded == nil || !app.Wanted {
		return nil
	}
	check, err := s.Check(ctx, db, loaded)
	if err != nil {
		return hperrors.Wrap(err)
	}
	if !check.ServiceGone {
		return nil
	}
	if err = s.Remove(ctx, db, loaded, false); err != nil {
		return hperrors.Wrap(err)
	}
	app.Recreated = true
	return nil
}

// SyncOutcome says what became of an app once the feature's apply ran,
// comparing it with what it was; and whether it is new, with nothing in it
// yet. tasks are the deployments the apply queued. Nil when there was no app
// and is none.
func SyncOutcome(ctx context.Context, db database.IDB, s Service, app *SyncedApp, tasks []*entity.Task) (
	*entity.SystemAppSyncOutput, bool, error) {
	after, err := s.LoadApp(ctx, db, app.Key)
	if err != nil {
		return nil, false, hperrors.Wrap(err)
	}
	out := &entity.SystemAppSyncOutput{Key: app.Key, Name: app.Name, Action: entity.SystemAppSyncNone}
	switch {
	case after == nil && app.Before == nil:
		return nil, false, nil
	case after == nil:
		out.AppID, out.Action = app.Before.ID, entity.SystemAppSyncRemoved
		out.Problem = "the settings no longer want it; its data was kept"
		return out, false, nil
	case app.Recreated:
		out.AppID, out.PreviousAppID, out.Action = after.ID, app.Before.ID, entity.SystemAppSyncRecreated
		out.Problem = "its service was gone; its data was kept"
		return out, true, nil
	case app.Before == nil:
		out.AppID, out.Action = after.ID, entity.SystemAppSyncProvisioned
		out.Problem = "the settings want it, and it was missing"
		return out, true, nil
	}
	out.AppID = after.ID
	if slices.ContainsFunc(tasks, func(t *entity.Task) bool { return t != nil && t.ObjectID == after.ID }) {
		out.Action, out.Problem = entity.SystemAppSyncUpdated, "it was deployed with what its settings say"
		return out, false, nil
	}
	check, err := s.Check(ctx, db, after)
	if err != nil {
		return nil, false, hperrors.Wrap(err)
	}
	out.Action, out.Problem = check.Action, check.Problem
	return out, false, nil
}
