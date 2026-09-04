package permissionimpl

import (
	"errors"
	"slices"
	"testing"
	"time"

	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
)

func actions(list ...base.ActionType) base.AccessActions {
	var out base.AccessActions
	for _, action := range list {
		switch action {
		case base.ActionTypeRead:
			out.Read = true
		case base.ActionTypeExecute:
			out.Exec = true
		case base.ActionTypeWrite:
			out.Write = true
		case base.ActionTypeDelete:
			out.Del = true
		}
	}
	return out
}

func envPerm(id string, list ...base.ActionType) *entity.ACLPermission {
	return envPermOf("usr_1", id, list...)
}

func envPermOf(subjectID, id string, list ...base.ActionType) *entity.ACLPermission {
	return &entity.ACLPermission{
		SubjectType:  base.SubjectTypeUser,
		SubjectID:    subjectID,
		ResourceType: base.ResourceTypeProjectEnv,
		ResourceID:   id,
		Actions:      actions(list...),
	}
}

func modulePerm(id string, list ...base.ActionType) *entity.ACLPermission {
	return &entity.ACLPermission{
		SubjectType:  base.SubjectTypeUser,
		ResourceType: base.ResourceTypeModule,
		ResourceID:   id,
		Actions:      actions(list...),
	}
}

func deleted(perm *entity.ACLPermission) *entity.ACLPermission {
	perm.DeletedAt = time.Now()
	return perm
}

// devEnvKey is the resource every table case grants on unless stated otherwise.
var devEnvKey = resourceKey{Type: base.ResourceTypeProjectEnv, ID: "prj_1:dev"}

func projectKey(id string) resourceKey {
	return resourceKey{Type: base.ResourceTypeProject, ID: id}
}

func moduleKey(id string) resourceKey {
	return resourceKey{Type: base.ResourceTypeModule, ID: id}
}

func TestCheckAccessChanges(t *testing.T) {
	tests := []struct {
		name    string
		actor   actorAccess
		desired []*entity.ACLPermission
		current []*entity.ACLPermission
		wantErr bool
		// wantReplaceable lists the resource IDs expected back; nil skips the check.
		wantReplaceable []string
	}{
		{
			name:    "granting what the actor holds is allowed",
			actor:   actorAccess{devEnvKey: actions(base.ActionTypeRead, base.ActionTypeWrite)},
			desired: []*entity.ACLPermission{envPerm("prj_1:dev", base.ActionTypeRead, base.ActionTypeWrite)},
		},
		{
			name:    "granting an action the actor lacks is rejected",
			actor:   actorAccess{devEnvKey: actions(base.ActionTypeRead)},
			desired: []*entity.ACLPermission{envPerm("prj_1:dev", base.ActionTypeRead, base.ActionTypeWrite)},
			wantErr: true,
		},
		{
			name:    "granting on a resource the actor has nothing on is rejected",
			actor:   actorAccess{},
			desired: []*entity.ACLPermission{envPerm("prj_1:dev", base.ActionTypeRead)},
			wantErr: true,
		},
		{
			// An admin may have granted more than the actor holds; leaving that
			// untouched must keep working, otherwise the actor cannot edit at all.
			name:    "an unchanged higher permission is left alone",
			actor:   actorAccess{devEnvKey: actions(base.ActionTypeRead)},
			desired: []*entity.ACLPermission{envPerm("prj_1:dev", base.ActionTypeRead, base.ActionTypeDelete)},
			current: []*entity.ACLPermission{envPerm("prj_1:dev", base.ActionTypeRead, base.ActionTypeDelete)},
			// ...and stays out of the replaceable set, so it survives the replace.
			wantReplaceable: []string{},
		},
		{
			name:    "revoking an action the actor holds is allowed",
			actor:   actorAccess{devEnvKey: actions(base.ActionTypeRead, base.ActionTypeWrite)},
			desired: []*entity.ACLPermission{envPerm("prj_1:dev", base.ActionTypeRead)},
			current: []*entity.ACLPermission{envPerm("prj_1:dev", base.ActionTypeRead, base.ActionTypeWrite)},
		},
		{
			name:    "revoking an action the actor lacks is rejected",
			actor:   actorAccess{devEnvKey: actions(base.ActionTypeRead)},
			desired: []*entity.ACLPermission{envPerm("prj_1:dev", base.ActionTypeRead)},
			current: []*entity.ACLPermission{envPerm("prj_1:dev", base.ActionTypeRead, base.ActionTypeDelete)},
			wantErr: true,
		},
		{
			name:    "an explicit revocation row is checked too",
			actor:   actorAccess{devEnvKey: actions(base.ActionTypeRead)},
			desired: []*entity.ACLPermission{deleted(envPerm("prj_1:dev", base.ActionTypeRead))},
			current: []*entity.ACLPermission{envPerm("prj_1:dev", base.ActionTypeRead, base.ActionTypeWrite)},
			wantErr: true,
		},
		{
			// Out of the actor's reach: it stays out of the replaceable set, so a
			// wholesale replace cannot silently revoke it.
			name:            "a row the actor cannot fully revoke is not replaceable",
			actor:           actorAccess{devEnvKey: actions(base.ActionTypeRead)},
			current:         []*entity.ACLPermission{envPerm("prj_1:dev", base.ActionTypeRead, base.ActionTypeWrite)},
			wantReplaceable: []string{},
		},
		{
			name:            "a row the actor fully controls is replaceable",
			actor:           actorAccess{devEnvKey: actions(base.ActionTypeRead, base.ActionTypeWrite)},
			current:         []*entity.ACLPermission{envPerm("prj_1:dev", base.ActionTypeRead, base.ActionTypeWrite)},
			wantReplaceable: []string{"prj_1:dev"},
		},
		{
			// The actor's project-level grant covers every env of that project.
			name:    "env access is inherited from the actor's project grant",
			actor:   actorAccess{projectKey("prj_1"): actions(base.ActionTypeRead, base.ActionTypeWrite)},
			desired: []*entity.ACLPermission{envPerm("prj_1:prod", base.ActionTypeRead, base.ActionTypeWrite)},
		},
		{
			name:    "a project grant of another project does not apply",
			actor:   actorAccess{projectKey("prj_2"): actions(base.ActionTypeRead, base.ActionTypeWrite)},
			desired: []*entity.ACLPermission{envPerm("prj_1:prod", base.ActionTypeRead)},
			wantErr: true,
		},
		{
			// The worst escalation path: handing out a module the actor lacks.
			name:    "granting a module the actor lacks is rejected",
			actor:   actorAccess{moduleKey("mod::user"): actions(base.ActionTypeRead, base.ActionTypeWrite)},
			desired: []*entity.ACLPermission{modulePerm("mod::system", base.ActionTypeWrite)},
			wantErr: true,
		},
		{
			// One call may carry several subjects: their rows must not be mixed up.
			name:  "changes are matched per subject",
			actor: actorAccess{devEnvKey: actions(base.ActionTypeRead)},
			desired: []*entity.ACLPermission{
				envPermOf("usr_1", "prj_1:dev", base.ActionTypeRead),
				envPermOf("usr_2", "prj_1:dev", base.ActionTypeRead),
			},
			// usr_2 already had write; usr_1 had nothing beyond read.
			current: []*entity.ACLPermission{
				envPermOf("usr_1", "prj_1:dev", base.ActionTypeRead),
				envPermOf("usr_2", "prj_1:dev", base.ActionTypeRead, base.ActionTypeWrite),
			},
			// Revoking usr_2's write is beyond the actor's reach.
			wantErr: true,
		},
		{
			name:    "granting a module the actor holds is allowed",
			actor:   actorAccess{moduleKey("mod::user"): actions(base.ActionTypeRead, base.ActionTypeWrite)},
			desired: []*entity.ACLPermission{modulePerm("mod::user", base.ActionTypeRead)},
			current: []*entity.ACLPermission{modulePerm("mod::user", base.ActionTypeRead, base.ActionTypeWrite)},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			replaceable, err := authorizeAccessChanges(tt.actor, tt.desired, tt.current)
			if tt.wantErr {
				if err == nil {
					t.Fatal("expected the change to be rejected, got nil")
				}
				if !errors.Is(err, hperrors.ErrForbidden) {
					t.Errorf("expected ErrForbidden, got %v", err)
				}
				return
			}
			if err != nil {
				t.Fatalf("expected the change to be allowed, got %v", err)
			}
			if tt.wantReplaceable == nil {
				return
			}
			got := make([]string, 0, len(replaceable))
			for _, perm := range replaceable {
				got = append(got, perm.ResourceID)
			}
			if !slices.Equal(got, tt.wantReplaceable) {
				t.Errorf("replaceable rows = %v, want %v", got, tt.wantReplaceable)
			}
		})
	}
}
