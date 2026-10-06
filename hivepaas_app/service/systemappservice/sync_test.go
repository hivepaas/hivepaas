package systemappservice

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/infra/database"
)

// apps is a Service holding apps by key, each checked as checks says.
type apps struct {
	Service
	byKey   map[string]*entity.App
	checks  map[string]*AppCheck
	removed []string
	storage bool
}

func (a *apps) LoadApp(_ context.Context, _ database.IDB, key string) (*entity.App, error) {
	return a.byKey[key], nil
}

func (a *apps) Check(_ context.Context, _ database.IDB, app *entity.App) (*AppCheck, error) {
	if check, ok := a.checks[app.Key]; ok {
		return check, nil
	}
	return &AppCheck{Action: entity.SystemAppSyncNone}, nil
}

func (a *apps) Remove(_ context.Context, _ database.IDB, app *entity.App, removeStorage bool) error {
	a.removed = append(a.removed, app.Key)
	a.storage = a.storage || removeStorage
	delete(a.byKey, app.Key)
	return nil
}

// Only an app the settings want, whose service is gone, is removed - its data
// kept - for the feature to provision it again.
func TestRemoveIfServiceGone(t *testing.T) {
	gone := &AppCheck{Action: entity.SystemAppSyncReported, ServiceGone: true}
	s := &apps{byKey: map[string]*entity.App{"a": {ID: "1", Key: "a"}, "b": {ID: "2", Key: "b"},
		"c": {ID: "3", Key: "c"}}, checks: map[string]*AppCheck{"a": gone, "b": gone}}

	wanted := &SyncedApp{Key: "a", Wanted: true}
	unwanted := &SyncedApp{Key: "b"}
	running := &SyncedApp{Key: "c", Wanted: true}
	missing := &SyncedApp{Key: "d", Wanted: true}
	for _, app := range []*SyncedApp{wanted, unwanted, running, missing} {
		assert.NoError(t, RemoveIfServiceGone(context.Background(), nil, s, app))
	}

	assert.Equal(t, []string{"a"}, s.removed)
	assert.False(t, s.storage, "its data stays")
	assert.True(t, wanted.Recreated)
	assert.Equal(t, "1", wanted.Before.ID)
	assert.False(t, unwanted.Recreated, "the feature removes what it does not want")
	assert.False(t, running.Recreated)
	assert.Nil(t, missing.Before)
}

// What became of an app, from what it was and is.
func TestSyncOutcome(t *testing.T) {
	before := &entity.App{ID: "1", Key: "a"}
	for _, tc := range []struct {
		name   string
		app    *SyncedApp
		after  *entity.App
		tasks  []*entity.Task
		check  *AppCheck
		action entity.SystemAppSyncAction
		fresh  bool
	}{
		{"removed", &SyncedApp{Key: "a", Before: before}, nil, nil, nil, entity.SystemAppSyncRemoved, false},
		{"provisioned", &SyncedApp{Key: "a", Wanted: true}, before, nil, nil, entity.SystemAppSyncProvisioned, true},
		{"recreated", &SyncedApp{Key: "a", Wanted: true, Before: before, Recreated: true},
			&entity.App{ID: "2", Key: "a"}, nil, nil, entity.SystemAppSyncRecreated, true},
		{"deployed", &SyncedApp{Key: "a", Wanted: true, Before: before}, before,
			[]*entity.Task{{ObjectID: "1"}}, nil, entity.SystemAppSyncUpdated, false},
		{"as it was", &SyncedApp{Key: "a", Wanted: true, Before: before}, before,
			[]*entity.Task{{ObjectID: "other"}}, nil, entity.SystemAppSyncNone, false},
		{"reported", &SyncedApp{Key: "a", Wanted: true, Before: before}, before, nil,
			&AppCheck{Action: entity.SystemAppSyncReported, Problem: "its service is scaled to zero"},
			entity.SystemAppSyncReported, false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			s := &apps{byKey: map[string]*entity.App{}, checks: map[string]*AppCheck{}}
			if tc.after != nil {
				s.byKey["a"] = tc.after
			}
			if tc.check != nil {
				s.checks["a"] = tc.check
			}
			out, fresh, err := SyncOutcome(context.Background(), nil, s, tc.app, tc.tasks)
			assert.NoError(t, err)
			if assert.NotNil(t, out) {
				assert.Equal(t, tc.action, out.Action)
			}
			assert.Equal(t, tc.fresh, fresh)
		})
	}

	out, _, err := SyncOutcome(context.Background(), nil, &apps{byKey: map[string]*entity.App{}},
		&SyncedApp{Key: "a"}, nil)
	assert.NoError(t, err)
	assert.Nil(t, out, "none was, none is")
}
