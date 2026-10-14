package scopeserviceimpl

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/uptrace/bun"
	"github.com/uptrace/bun/dialect/pgdialect"

	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/infra/database"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/bunex"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/appservice"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/projectservice"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/userservice"
)

// queries keeps, in order, the SQL each load would have run.
type queries struct{ sql []string }

func (q *queries) add(model any, opts []bunex.SelectQueryOption) {
	query := bun.NewDB(nil, pgdialect.New()).NewSelect().Model(model)
	q.sql = append(q.sql, bunex.ApplySelect(query, opts...).String())
}

type queriedProjects struct {
	projectservice.Service
	*queries
}

func (f queriedProjects) LoadProject(_ context.Context, _ database.IDB, id string, _ bool,
	opts ...bunex.SelectQueryOption) (*entity.Project, error) {
	f.add((*entity.Project)(nil), opts)
	return &entity.Project{ID: id}, nil
}

func (f queriedProjects) LoadProjectEnv(_ context.Context, _ database.IDB, projectID, envID string, _, _ bool,
	opts ...bunex.SelectQueryOption) (*entity.ProjectEnv, error) {
	f.add((*entity.ProjectEnv)(nil), opts)
	return &entity.ProjectEnv{ID: envID, ProjectID: projectID, Project: &entity.Project{ID: projectID}}, nil
}

type queriedApps struct {
	appservice.Service
	*queries
}

func (f queriedApps) LoadApp(_ context.Context, _ database.IDB, projectID, appID string, _, _ bool,
	opts ...bunex.SelectQueryOption) (*entity.App, error) {
	f.add((*entity.App)(nil), opts)
	return &entity.App{ID: appID, ProjectID: projectID, ProjectEnvID: projectID + ":dev"}, nil
}

type queriedUsers struct {
	userservice.Service
	*queries
}

func (f queriedUsers) LoadUser(_ context.Context, _ database.IDB, id string, _ bool,
	opts ...bunex.SelectQueryOption) (*entity.User, error) {
	f.add((*entity.User)(nil), opts)
	return &entity.User{ID: id}, nil
}

func newQueriedService() (*service, *queries) {
	q := &queries{}
	return &service{projectService: queriedProjects{queries: q}, appService: queriedApps{queries: q},
		userService: queriedUsers{queries: q}}, q
}

// A locked app's scope holds its project, its environment and itself against
// deletion, in that order - the order the deletions lock them in - each with a
// key-share lock, which other writes pass.
func TestALockedAppScopeHoldsItsProjectEnvironmentAndItself(t *testing.T) {
	t.Parallel()
	s, q := newQueriedService()
	scope := entity.NewObjectScopeApp("a1", "", "p1", "p1:dev")
	scope.LockScopeObject = true

	assert.NoError(t, s.LoadObjectScopeData(context.Background(), nil, scope))

	if assert.Len(t, q.sql, 3) {
		assert.Contains(t, q.sql[0], "FOR KEY SHARE OF project")
		assert.Contains(t, q.sql[1], "FOR KEY SHARE OF project_env")
		assert.Contains(t, q.sql[2], "FOR KEY SHARE OF app")
	}
}

func TestALockedEnvironmentScopeHoldsItsProjectAndItself(t *testing.T) {
	t.Parallel()
	s, q := newQueriedService()
	scope := entity.NewObjectScopeProjectEnv("p1", "p1:dev")
	scope.LockScopeObject = true

	assert.NoError(t, s.LoadObjectScopeData(context.Background(), nil, scope))

	if assert.Len(t, q.sql, 2) {
		assert.Contains(t, q.sql[0], "FOR KEY SHARE OF project")
		assert.Contains(t, q.sql[1], "FOR KEY SHARE OF project_env")
	}
}

// "user" is a reserved word: unquoted, the lock is a syntax error.
func TestALockedProjectOrUserScopeHoldsItsObject(t *testing.T) {
	t.Parallel()
	s, q := newQueriedService()
	project := entity.NewObjectScopeProject("p1")
	project.LockScopeObject = true
	user := entity.NewObjectScopeUser("u1")
	user.LockScopeObject = true

	assert.NoError(t, s.LoadObjectScopeData(context.Background(), nil, project))
	assert.NoError(t, s.LoadObjectScopeData(context.Background(), nil, user))

	if assert.Len(t, q.sql, 2) {
		assert.Contains(t, q.sql[0], "FOR KEY SHARE OF project")
		assert.Contains(t, q.sql[1], `FOR KEY SHARE OF "user"`)
	}
}

// Unlocked, a scope is read as before: one query, no lock - and no lock
// written into its WHERE, as the lock once was.
func TestAnUnlockedScopeTakesNoLock(t *testing.T) {
	t.Parallel()
	s, q := newQueriedService()
	scope := entity.NewObjectScopeApp("a1", "", "p1", "p1:dev")

	assert.NoError(t, s.LoadObjectScopeData(context.Background(), nil, scope))

	if assert.Len(t, q.sql, 1) {
		assert.NotContains(t, q.sql[0], " FOR ")
		assert.NotContains(t, q.sql[0], "SHARE")
	}
	assert.Equal(t, base.ObjectScopeApp, scope.ScopeType)
}
