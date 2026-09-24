package permissionimpl

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/basedto"
	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/infra/database"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/bunex"
	"github.com/hivepaas/hivepaas/hivepaas_app/repository"
)

// rowsACLRepo answers from rows held in memory, matching them the way the real
// query does: by resource id, or by resource type when the id is empty.
type rowsACLRepo struct {
	repository.ACLPermissionRepo
	rows  []*entity.ACLPermission
	calls int
}

func (f *rowsACLRepo) ListByResources(
	_ context.Context, _ database.IDB, resources []*base.PermissionResource, _ ...bunex.SelectQueryOption,
) ([]*entity.ACLPermission, error) {
	f.calls++
	var found []*entity.ACLPermission
	for _, row := range f.rows {
		for _, res := range resources {
			if row.SubjectID != res.SubjectID {
				continue
			}
			if (res.ResourceID != "" && row.ResourceID == res.ResourceID) ||
				(res.ResourceID == "" && row.ResourceType == res.ResourceType) {
				found = append(found, row)
				break
			}
		}
	}
	return found, nil
}

// ownersProjectRepo knows which user owns which project, and nothing else.
type ownersProjectRepo struct {
	repository.ProjectRepo
	owners map[string]string
}

func (f *ownersProjectRepo) GetByIDAndOwner(
	_ context.Context, _ database.IDB, projectID, ownerID string, _ ...bunex.SelectQueryOption,
) (*entity.Project, error) {
	if f.owners[projectID] == ownerID {
		return &entity.Project{ID: projectID}, nil
	}
	return nil, hperrors.ErrNotFound
}

// List answers the one list the manager makes, a user's own projects. Every
// test asks as usr_1, so those are the ones returned.
func (f *ownersProjectRepo) List(
	_ context.Context, _ database.IDB, _ *basedto.Paging, _ ...bunex.SelectQueryOption,
) ([]*entity.Project, *basedto.PagingMeta, error) {
	var projects []*entity.Project
	for projectID, ownerID := range f.owners {
		if ownerID == "usr_1" {
			projects = append(projects, &entity.Project{ID: projectID})
		}
	}
	return projects, nil, nil
}

func grant(resType base.ResourceType, resID string, actions base.AccessActions) *entity.ACLPermission {
	return &entity.ACLPermission{
		SubjectType: base.SubjectTypeUser, SubjectID: "usr_1",
		ResourceType: resType, ResourceID: resID, Actions: actions,
	}
}

var readOnly = base.AccessActions{Read: true}

func newVisibility(rows []*entity.ACLPermission, owners map[string]string) (*visibility, *rowsACLRepo, *basedto.Auth) {
	acl := &rowsACLRepo{rows: rows}
	p := &manager{aclPermissionRepo: acl, projectRepo: &ownersProjectRepo{owners: owners}}
	auth := plainAuth()
	return p.NewVisibility(nil, auth).(*visibility), acl, auth
}

// answered turns an answer and its error into the answer, failing the test on
// the error.
func answered(t *testing.T) func(bool, error) bool {
	t.Helper()
	return func(answer bool, err error) bool {
		t.Helper()
		assert.NoError(t, err)
		return answer
	}
}

// An admin is let in everywhere without anything being looked up.
func TestVisibilityLetsAnAdminSeeEverything(t *testing.T) {
	allows := answered(t)
	p := &manager{} // no repositories: a lookup would panic
	v := p.NewVisibility(nil, adminAuth())
	ctx := context.Background()

	assert.True(t, allows(v.AllowsModule(ctx, base.ResourceModuleCluster, base.ActionTypeRead)))
	assert.True(t, allows(v.AllowsProjectEnv(ctx, "prj_1", "prod", base.ActionTypeWrite)))
}

// A module grant opens that module only, and for the actions it names.
func TestVisibilityAnswersModulesByTheirOwnGrant(t *testing.T) {
	allows := answered(t)
	v, _, _ := newVisibility([]*entity.ACLPermission{
		grant(base.ResourceTypeModule, string(base.ResourceModuleCluster), readOnly),
	}, nil)
	ctx := context.Background()

	assert.True(t, allows(v.AllowsModule(ctx, base.ResourceModuleCluster, base.ActionTypeRead)))
	assert.False(t, allows(v.AllowsModule(ctx, base.ResourceModuleCluster, base.ActionTypeWrite)))
	assert.False(t, allows(v.AllowsModule(ctx, base.ResourceModuleSystem, base.ActionTypeRead)))
}

// A grant on a project reaches its envs, and an env's own row - an env the
// user is kept out of - decides for that env.
func TestVisibilityLetsAnEnvsGrantDecideOverItsProjects(t *testing.T) {
	allows := answered(t)
	v, _, _ := newVisibility([]*entity.ACLPermission{
		grant(base.ResourceTypeProject, "prj_1", readOnly),
		grant(base.ResourceTypeProjectEnv, "prj_1:prod", base.AccessActions{}),
	}, nil)
	ctx := context.Background()

	assert.True(t, allows(v.AllowsProjectEnv(ctx, "prj_1", "dev", base.ActionTypeRead)))
	assert.False(t, allows(v.AllowsProjectEnv(ctx, "prj_1", "prod", base.ActionTypeRead)))
	assert.False(t, allows(v.AllowsProjectEnv(ctx, "prj_2", "dev", base.ActionTypeRead)), "another project")
}

// An owner holds every action on their project with no grant written for it.
func TestVisibilityLetsAnOwnerIntoTheirProject(t *testing.T) {
	allows := answered(t)
	v, _, _ := newVisibility(nil, map[string]string{"prj_1": "usr_1"})
	ctx := context.Background()

	assert.True(t, allows(v.AllowsProjectEnv(ctx, "prj_1", "prod", base.ActionTypeDelete)))
	assert.False(t, allows(v.AllowsProjectEnv(ctx, "prj_2", "prod", base.ActionTypeRead)))
}

// Asking about a whole project would answer yes for a grant on any one of its
// envs, so a question without an env is refused rather than asked.
func TestVisibilityNeedsAnEnv(t *testing.T) {
	allows := answered(t)
	v, acl, _ := newVisibility([]*entity.ACLPermission{
		grant(base.ResourceTypeProjectEnv, "prj_1:dev", readOnly),
	}, nil)

	assert.False(t, allows(v.AllowsProjectEnv(context.Background(), "prj_1", "", base.ActionTypeRead)))
	assert.Zero(t, acl.calls)
}

// One request asks about the same env for many items: it is looked up once,
// and nothing is left on the caller's auth for its own lists to read.
func TestVisibilityAsksOnceAndLeavesTheCallersAuthAlone(t *testing.T) {
	allows := answered(t)
	v, acl, auth := newVisibility([]*entity.ACLPermission{
		grant(base.ResourceTypeProject, "prj_1", readOnly),
	}, nil)
	ctx := context.Background()

	for range 3 {
		assert.True(t, allows(v.AllowsProjectEnv(ctx, "prj_1", "dev", base.ActionTypeRead)))
	}
	assert.False(t, allows(v.AllowsProjectEnv(ctx, "prj_1", "dev", base.ActionTypeWrite)))

	assert.Equal(t, 2, acl.calls, "once per env and action")
	assert.Nil(t, auth.AllowedResources)
}
