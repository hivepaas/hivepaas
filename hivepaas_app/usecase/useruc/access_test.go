package useruc

import (
	"slices"
	"testing"

	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/basedto"
	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/timeutil"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/userservice"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/useruc/userdto"
)

func TestPreparePersistingUserProjectAccessesWritesEnvLevelRows(t *testing.T) {
	uc := &UC{}
	user := &entity.User{ID: "usr_1"}
	persistingData := &userservice.PersistingUserData{}

	uc.preparePersistingUserProjectAccesses(user, []*userdto.ProjectAccessReq{
		{
			Project: basedto.ObjectIDReq{ID: "prj_1"},
			EnvAccesses: basedto.ObjectAccessSliceReq{
				{ObjectIDReq: basedto.ObjectIDReq{ID: "prj_1:dev"},
					Access: base.AccessActions{Read: true, Write: true}},
				{ObjectIDReq: basedto.ObjectIDReq{ID: "prj_1:prod"},
					Access: base.AccessActions{Read: true}},
			},
		},
	}, timeutil.NowUTC(), persistingData)

	if len(persistingData.UpsertingAccesses) != 2 {
		t.Fatalf("expected 2 ACL rows, got %d", len(persistingData.UpsertingAccesses))
	}
	for _, acl := range persistingData.UpsertingAccesses {
		if acl.ResourceType != base.ResourceTypeProjectEnv {
			t.Errorf("row %q should be a project-env grant, got %q", acl.ResourceID, acl.ResourceType)
		}
		if acl.SubjectType != base.SubjectTypeUser || acl.SubjectID != user.ID {
			t.Errorf("row %q has the wrong subject: %s/%s", acl.ResourceID, acl.SubjectType, acl.SubjectID)
		}
	}
	if got := persistingData.UpsertingAccesses[0]; got.ResourceID != "prj_1:dev" || !got.Actions.Write {
		t.Errorf("first row should keep its ID and actions, got %+v", got)
	}
}

func TestAccessResourceTypesToReplace(t *testing.T) {
	tests := []struct {
		name             string
		replaceModules   bool
		replaceProjects  bool
		wantResourceType []base.ResourceType
	}{
		{
			name: "nothing to replace when the request carries no access list",
		},
		{
			name:             "modules only",
			replaceModules:   true,
			wantResourceType: []base.ResourceType{base.ResourceTypeModule},
		},
		{
			// The legacy project level must be replaced too, otherwise an old row
			// would linger and keep granting access to every env of the project.
			name:            "projects cover both the env and the legacy project level",
			replaceProjects: true,
			wantResourceType: []base.ResourceType{
				base.ResourceTypeProject, base.ResourceTypeProjectEnv,
			},
		},
		{
			name:            "both",
			replaceModules:  true,
			replaceProjects: true,
			wantResourceType: []base.ResourceType{
				base.ResourceTypeModule, base.ResourceTypeProject, base.ResourceTypeProjectEnv,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := accessResourceTypesToReplace(tt.replaceModules, tt.replaceProjects)
			if !slices.Equal(got, tt.wantResourceType) {
				t.Errorf("accessResourceTypesToReplace() = %v, want %v", got, tt.wantResourceType)
			}
		})
	}
}
