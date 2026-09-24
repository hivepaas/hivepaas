package permissionimpl

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/basedto"
	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/permission"
)

var fullAccess = base.AccessActions{Read: true, Exec: true, Write: true, Del: true}

func newProjectManager(rows []*entity.ACLPermission, owners map[string]string) *manager {
	return &manager{aclPermissionRepo: &rowsACLRepo{rows: rows}, projectRepo: &ownersProjectRepo{owners: owners}}
}

// listProjects asks what GET /projects asks, and returns what its list is then
// narrowed to: every project, some, or - on a refusal - nothing at all.
func listProjects(t *testing.T, p *manager) (allowed bool, all bool, projectIDs []string) {
	t.Helper()
	auth := plainAuth()
	allowed, err := p.CheckAccess(context.Background(), nil, auth, &permission.ProjectAccessCheck{
		BaseAccessCheck: permission.BaseAccessCheck{Action: base.ActionTypeRead},
	})
	assert.NoError(t, err)
	all, projectIDs = auth.AllowedProjects(nil)
	return allowed, all, projectIDs
}

func checkProject(t *testing.T, p *manager, projectID string, env *string, action base.ActionType) bool {
	t.Helper()
	allowed, err := p.CheckAccess(context.Background(), nil, plainAuth(), &permission.ProjectAccessCheck{
		BaseAccessCheck: permission.BaseAccessCheck{Action: action},
		ProjectID:       projectID,
		ProjectEnv:      env,
	})
	assert.NoError(t, err)
	return allowed
}

func TestProjectListFollowsTheGrants(t *testing.T) {
	tests := []struct {
		name     string
		rows     []*entity.ACLPermission
		owners   map[string]string
		allowed  bool
		all      bool
		projects []string
	}{
		{
			name:    "the module grant opens every project",
			rows:    []*entity.ACLPermission{grant(base.ResourceTypeModule, string(base.ResourceModuleProject), readOnly)},
			allowed: true, all: true,
		},
		{
			name:     "a grant on one env lists its project",
			rows:     []*entity.ACLPermission{grant(base.ResourceTypeProjectEnv, "prj_1:dev", readOnly)},
			allowed:  true,
			projects: []string{"prj_1"},
		},
		{
			name:     "a grant on the project lists it",
			rows:     []*entity.ACLPermission{grant(base.ResourceTypeProject, "prj_1", readOnly)},
			allowed:  true,
			projects: []string{"prj_1"},
		},
		{
			name: "a module row without reading does not hide what was granted inside",
			rows: []*entity.ACLPermission{
				grant(base.ResourceTypeModule, string(base.ResourceModuleProject), base.AccessActions{}),
				grant(base.ResourceTypeProjectEnv, "prj_2:prod", readOnly),
			},
			allowed:  true,
			projects: []string{"prj_2"},
		},
		{
			name:     "an owner sees their project without a grant",
			owners:   map[string]string{"prj_3": "usr_1"},
			allowed:  true,
			projects: []string{"prj_3"},
		},
		{
			name: "a row that grants nothing lists nothing",
			rows: []*entity.ACLPermission{
				grant(base.ResourceTypeProject, "prj_1", base.AccessActions{}),
				grant(base.ResourceTypeProjectEnv, "prj_2:dev", base.AccessActions{}),
			},
		},
		{
			name: "nothing at all is refused",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			allowed, all, projects := listProjects(t, newProjectManager(tt.rows, tt.owners))

			assert.Equal(t, tt.allowed, allowed)
			if tt.allowed {
				assert.Equal(t, tt.all, all)
				assert.ElementsMatch(t, tt.projects, projects)
			}
		})
	}
}

// A user given one env opens the project it is in, and only that project.
func TestProjectOpensForAGrantOnOneOfItsEnvs(t *testing.T) {
	p := newProjectManager([]*entity.ACLPermission{grant(base.ResourceTypeProjectEnv, "prj_1:dev", readOnly)}, nil)
	prod := "prod"

	assert.True(t, checkProject(t, p, "prj_1", nil, base.ActionTypeRead))
	assert.False(t, checkProject(t, p, "prj_1", &prod, base.ActionTypeRead), "another env of it")
	assert.False(t, checkProject(t, p, "prj_2", nil, base.ActionTypeRead), "another project")
}

// Changing or deleting a project reaches every env of it, so full access to
// one env is not enough - nor is it enough to make a project.
func TestAGrantOnOneEnvDoesNotReachTheWholeProject(t *testing.T) {
	p := newProjectManager([]*entity.ACLPermission{grant(base.ResourceTypeProjectEnv, "prj_1:dev", fullAccess)}, nil)
	dev := "dev"

	assert.True(t, checkProject(t, p, "prj_1", &dev, base.ActionTypeDelete))
	assert.False(t, checkProject(t, p, "prj_1", nil, base.ActionTypeWrite))
	assert.False(t, checkProject(t, p, "prj_1", nil, base.ActionTypeDelete))
	assert.False(t, checkProject(t, p, "", nil, base.ActionTypeWrite), "making a project is the module's to allow")
}

// An env a user is kept out of stays shut, whatever else is granted.
func TestAnEnvWithoutAGrantStaysShut(t *testing.T) {
	p := newProjectManager([]*entity.ACLPermission{
		grant(base.ResourceTypeModule, string(base.ResourceModuleProject), readOnly),
		grant(base.ResourceTypeProjectEnv, "prj_1:prod", base.AccessActions{}),
	}, nil)
	prod, dev := "prod", "dev"

	assert.False(t, checkProject(t, p, "prj_1", &prod, base.ActionTypeRead))
	assert.True(t, checkProject(t, p, "prj_1", &dev, base.ActionTypeRead))
}

// A list names each project once, however many of its envs were granted.
func TestAllowedIDsAreEachListedOnce(t *testing.T) {
	auth := &basedto.Auth{AllowedResources: map[base.ResourceType][]string{
		base.ResourceTypeProject: {"prj_1", "prj_2", "prj_1"},
	}}

	all, ids := auth.AllowedProjects(nil)

	assert.False(t, all)
	assert.ElementsMatch(t, []string{"prj_1", "prj_2"}, ids)
}
