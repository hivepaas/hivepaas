package useruc

import (
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

// Editing a user must also clear the legacy project-level rows, otherwise they
// would keep granting access to every env of the project.
func TestDeletingUserProjectAccessesCoversBothLevels(t *testing.T) {
	resources := deletingUserProjectAccesses("usr_1")

	seen := map[base.ResourceType]bool{}
	for _, res := range resources {
		if res.SubjectType != base.SubjectTypeUser || res.SubjectID != "usr_1" {
			t.Errorf("wrong subject: %s/%s", res.SubjectType, res.SubjectID)
		}
		seen[res.ResourceType] = true
	}
	for _, want := range []base.ResourceType{base.ResourceTypeProjectEnv, base.ResourceTypeProject} {
		if !seen[want] {
			t.Errorf("resource type %q should be cleared", want)
		}
	}
}
