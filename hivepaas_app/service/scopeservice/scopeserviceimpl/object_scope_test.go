package scopeserviceimpl

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/infra/database"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/bunex"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/appservice"
)

// A scope type that is not in the enum reaches LoadObjectScope whenever it comes from stored data
// instead of a validated request. It has to come back as an error, not a nil-pointer panic.
func TestLoadObjectScope_UnknownScopeType(t *testing.T) {
	t.Parallel()

	s := &service{} // the unknown-type path must fail before it touches any dependency

	scope, err := s.LoadObjectScope(context.Background(), nil, base.ObjectScopeType("bogus"), "id-1", true)

	assert.Error(t, err)
	assert.Nil(t, scope)
}

// appsByID finds the apps it holds, and none other.
type appsByID struct {
	appservice.Service
	apps map[string]*entity.App
}

func (f *appsByID) LoadApp(
	_ context.Context, _ database.IDB, _, appID string, _, _ bool, _ ...bunex.SelectQueryOption,
) (*entity.App, error) {
	if app, ok := f.apps[appID]; ok {
		return app, nil
	}
	return nil, hperrors.Wrap(hperrors.ErrAppNotFound).WithParam("Name", appID)
}

// A filter by app is the app's scope, found where it is; within a project, an
// app of another project, or none at all, filters to nothing.
func TestFilterScope_ByApp(t *testing.T) {
	t.Parallel()
	s := &service{appService: &appsByID{apps: map[string]*entity.App{
		"a1": {ID: "a1", ProjectID: "p1", ProjectEnvID: "p1:dev"},
		"a2": {ID: "a2", ProjectID: "p2", ProjectEnvID: "p2:dev"},
	}}}
	project := entity.NewObjectScopeProject("p1")
	ctx := context.Background()

	scope, err := s.FilterScope(ctx, nil, project, "", "", "a1")
	assert.NoError(t, err)
	assert.Equal(t, entity.NewObjectScopeApp("a1", "", "p1", "p1:dev"), scope)

	scope, err = s.FilterScope(ctx, nil, project, "", "", "a2")
	assert.NoError(t, err)
	assert.Nil(t, scope, "another project's app")

	scope, err = s.FilterScope(ctx, nil, project, "", "", "a9")
	assert.NoError(t, err)
	assert.Nil(t, scope, "no such app")

	scope, err = s.FilterScope(ctx, nil, project, "p2", "", "")
	assert.NoError(t, err)
	assert.Nil(t, scope, "another project")
}
