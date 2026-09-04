package userdto

import (
	"testing"

	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
)

func rw() base.AccessActions   { return base.AccessActions{Read: true, Write: true} }
func read() base.AccessActions { return base.AccessActions{Read: true} }

func newProject(id, name string, envNames ...string) *entity.Project {
	project := &entity.Project{ID: id, Name: name}
	for idx, envName := range envNames {
		project.ProjectEnvs = append(project.ProjectEnvs, &entity.ProjectEnv{
			ID:        id + ":" + envName,
			ProjectID: id,
			Name:      envName,
			Color:     "#" + envName,
			Index:     idx,
		})
	}
	return project
}

func envOf(project *entity.Project, name string) *entity.ProjectEnv {
	for _, env := range project.ProjectEnvs {
		if env.Name == name {
			env.Project = project
			return env
		}
	}
	return nil
}

// gotEnvs flattens one project's env accesses into name -> actions.
func gotEnvs(t *testing.T, resp *UserDetailsResp, projectName string) map[string]base.AccessActions {
	t.Helper()
	for _, projectAccess := range resp.ProjectAccesses {
		if projectAccess.Project.Name != projectName {
			continue
		}
		out := make(map[string]base.AccessActions, len(projectAccess.EnvAccesses))
		for _, envAccess := range projectAccess.EnvAccesses {
			out[envAccess.Name] = envAccess.Access
		}
		return out
	}
	t.Fatalf("project %q missing from the response", projectName)
	return nil
}

func TestTransformUserDetailsProjectAccessAppliesToEveryEnv(t *testing.T) {
	shop := newProject("prj_1", "Shop", "dev", "prod")
	user := &entity.User{
		ID: "usr_1",
		Accesses: []*entity.ACLPermission{
			{ResourceType: base.ResourceTypeProject, ResourceID: shop.ID,
				ResourceProject: shop, Actions: rw()},
		},
	}

	resp, err := TransformUserDetails(user)
	if err != nil {
		t.Fatalf("TransformUserDetails: %v", err)
	}

	envs := gotEnvs(t, resp, "Shop")
	if len(envs) != 2 {
		t.Fatalf("expected both envs to be listed, got %v", envs)
	}
	for _, name := range []string{"dev", "prod"} {
		if envs[name] != rw() {
			t.Errorf("env %q should inherit the project access, got %+v", name, envs[name])
		}
	}
}

func TestTransformUserDetailsEnvAccessOverridesProjectAccess(t *testing.T) {
	shop := newProject("prj_1", "Shop", "dev", "prod")
	user := &entity.User{
		ID: "usr_1",
		Accesses: []*entity.ACLPermission{
			{ResourceType: base.ResourceTypeProject, ResourceID: shop.ID,
				ResourceProject: shop, Actions: rw()},
			{ResourceType: base.ResourceTypeProjectEnv, ResourceID: "prj_1:prod",
				ResourceProjectEnv: envOf(shop, "prod"), Actions: read()},
		},
	}

	resp, err := TransformUserDetails(user)
	if err != nil {
		t.Fatalf("TransformUserDetails: %v", err)
	}

	envs := gotEnvs(t, resp, "Shop")
	if envs["dev"] != rw() {
		t.Errorf("dev should inherit the project access, got %+v", envs["dev"])
	}
	if envs["prod"] != read() {
		t.Errorf("prod should use its own access, got %+v", envs["prod"])
	}
}

// A user granted only one env must still see the other envs, with no access, so
// the caller can render and edit the whole matrix.
func TestTransformUserDetailsEnvOnlyAccessListsEveryEnv(t *testing.T) {
	shop := newProject("prj_1", "Shop", "dev", "prod")
	user := &entity.User{
		ID: "usr_1",
		Accesses: []*entity.ACLPermission{
			{ResourceType: base.ResourceTypeProjectEnv, ResourceID: "prj_1:dev",
				ResourceProjectEnv: envOf(shop, "dev"), Actions: rw()},
		},
	}

	resp, err := TransformUserDetails(user)
	if err != nil {
		t.Fatalf("TransformUserDetails: %v", err)
	}
	if len(resp.ProjectAccesses) != 1 {
		t.Fatalf("expected 1 project, got %d", len(resp.ProjectAccesses))
	}

	envs := gotEnvs(t, resp, "Shop")
	if envs["dev"] != rw() {
		t.Errorf("dev should keep its own access, got %+v", envs["dev"])
	}
	noAccess, ok := envs["prod"]
	if !ok {
		t.Fatal("prod should be listed even without any access")
	}
	if !noAccess.IsNoAccess() {
		t.Errorf("prod should report no access, got %+v", noAccess)
	}
}

func TestTransformUserDetailsEnvsKeepConfiguredOrderAndProjectsSortByName(t *testing.T) {
	// Envs are declared out of alphabetical order on purpose.
	shop := newProject("prj_1", "Shop", "prod", "dev")
	blog := newProject("prj_2", "Blog", "dev")
	user := &entity.User{
		ID: "usr_1",
		Accesses: []*entity.ACLPermission{
			{ResourceType: base.ResourceTypeProject, ResourceID: shop.ID,
				ResourceProject: shop, Actions: rw()},
			{ResourceType: base.ResourceTypeProject, ResourceID: blog.ID,
				ResourceProject: blog, Actions: read()},
			{ResourceType: base.ResourceTypeModule, ResourceID: "projects", Actions: read()},
		},
	}

	resp, err := TransformUserDetails(user)
	if err != nil {
		t.Fatalf("TransformUserDetails: %v", err)
	}

	got := []string{resp.ProjectAccesses[0].Project.Name, resp.ProjectAccesses[1].Project.Name}
	if got[0] != "Blog" || got[1] != "Shop" {
		t.Errorf("projects should be sorted by name, got %v", got)
	}
	shopEnvs := resp.ProjectAccesses[1].EnvAccesses
	if shopEnvs[0].Name != "prod" || shopEnvs[1].Name != "dev" {
		t.Errorf("envs should keep their configured index order, got %q, %q",
			shopEnvs[0].Name, shopEnvs[1].Name)
	}
	if len(resp.ModuleAccesses) != 1 || resp.ModuleAccesses[0].ID != "projects" {
		t.Errorf("module accesses should be unchanged, got %+v", resp.ModuleAccesses)
	}
}

// The env's project relation may not be loaded; the access must not be dropped.
func TestTransformUserDetailsKeepsAccessWhenProjectRelationMissing(t *testing.T) {
	user := &entity.User{
		ID: "usr_1",
		Accesses: []*entity.ACLPermission{
			{ResourceType: base.ResourceTypeProjectEnv, ResourceID: "prj_1:dev",
				ResourceProjectEnv: &entity.ProjectEnv{
					ID: "prj_1:dev", ProjectID: "prj_1", Name: "dev",
				},
				Actions: rw()},
		},
	}

	resp, err := TransformUserDetails(user)
	if err != nil {
		t.Fatalf("TransformUserDetails: %v", err)
	}
	if len(resp.ProjectAccesses) != 1 {
		t.Fatalf("expected the access to be kept, got %d projects", len(resp.ProjectAccesses))
	}
	if got := resp.ProjectAccesses[0].Project.ID; got != "prj_1" {
		t.Errorf("project ID should come from the env, got %q", got)
	}
	if resp.ProjectAccesses[0].EnvAccesses[0].Access != rw() {
		t.Error("env access should be preserved")
	}
}

// The env color travels with the access so the caller can render the env badge.
func TestTransformUserDetailsCarriesEnvColor(t *testing.T) {
	shop := newProject("prj_1", "Shop", "dev")
	user := &entity.User{
		ID: "usr_1",
		Accesses: []*entity.ACLPermission{
			{ResourceType: base.ResourceTypeProject, ResourceID: shop.ID,
				ResourceProject: shop, Actions: rw()},
		},
	}

	resp, err := TransformUserDetails(user)
	if err != nil {
		t.Fatalf("TransformUserDetails: %v", err)
	}
	if got := resp.ProjectAccesses[0].EnvAccesses[0].Color; got != "#dev" {
		t.Errorf("env color should be reported, got %q", got)
	}
}
