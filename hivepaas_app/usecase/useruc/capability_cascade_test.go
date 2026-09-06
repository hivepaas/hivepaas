package useruc

import (
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/basedto"
	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/infra/database"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/bunex"
	"github.com/hivepaas/hivepaas/hivepaas_app/repository"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/auditservice"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/userservice"
)

func aclRow(resType base.ResourceType, resID string) *entity.ACLPermission {
	return &entity.ACLPermission{
		SubjectType: base.SubjectTypeUser, SubjectID: "usr_1",
		ResourceType: resType, ResourceID: resID,
		Actions: base.AccessActions{Exec: true},
	}
}

func permResource(resType base.ResourceType, resID string) *base.PermissionResource {
	return &base.PermissionResource{
		SubjectType: base.SubjectTypeUser, SubjectID: "usr_1",
		ResourceType: resType, ResourceID: resID,
	}
}

func TestRevokedCapabilities(t *testing.T) {
	apiKeyCap := string(base.ResourceCapAPIKeyCreate)
	revealCap := string(base.ResourceCapSecretReveal)

	tests := []struct {
		name      string
		upserting []*entity.ACLPermission
		deleting  []*base.PermissionResource
		want      []base.ResourceCapability
	}{
		{
			name:     "a capability deleted and not written back is revoked",
			deleting: []*base.PermissionResource{permResource(base.ResourceTypeCapability, apiKeyCap)},
			want:     []base.ResourceCapability{base.ResourceCapAPIKeyCreate},
		},
		{
			// A wholesale replace deletes every row it is about to write again, so
			// a capability the request still asks for appears in both lists. Reading
			// the deletions alone would revoke keys on an update that changed nothing.
			name:      "a capability replaced by itself is not revoked",
			upserting: []*entity.ACLPermission{aclRow(base.ResourceTypeCapability, apiKeyCap)},
			deleting:  []*base.PermissionResource{permResource(base.ResourceTypeCapability, apiKeyCap)},
		},
		{
			name:      "only the capability that actually goes is reported",
			upserting: []*entity.ACLPermission{aclRow(base.ResourceTypeCapability, revealCap)},
			deleting: []*base.PermissionResource{
				permResource(base.ResourceTypeCapability, revealCap),
				permResource(base.ResourceTypeCapability, apiKeyCap),
			},
			want: []base.ResourceCapability{base.ResourceCapAPIKeyCreate},
		},
		{
			name: "module and project rows are not capabilities",
			deleting: []*base.PermissionResource{
				permResource(base.ResourceTypeModule, string(base.ResourceModuleUser)),
				permResource(base.ResourceTypeProjectEnv, "prj_1:dev"),
			},
		},
		{
			name: "nothing deleted, nothing revoked",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := revokedCapabilities(&userservice.PersistingUserData{
				UpsertingAccesses: tt.upserting,
				DeletingAccesses:  tt.deleting,
			})
			if len(got) != len(tt.want) {
				t.Fatalf("revokedCapabilities() = %v, want %v", got, tt.want)
			}
			for i := range got {
				if got[i] != tt.want[i] {
					t.Errorf("revokedCapabilities()[%d] = %q, want %q", i, got[i], tt.want[i])
				}
			}
		})
	}
}

/// Fakes for the cascade itself

type fakeSettingRepo struct {
	repository.SettingRepo
	settings []*entity.Setting
}

func (f *fakeSettingRepo) List(_ context.Context, _ database.IDB, _ *entity.ObjectScope,
	_ *basedto.Paging, _ ...bunex.SelectQueryOption) (
	[]*entity.Setting, *basedto.PagingMeta, error) {
	return f.settings, nil, nil
}

type fakeAuditService struct {
	entries []*auditservice.Entry
	err     error
}

func (f *fakeAuditService) Record(_ context.Context, _ database.IDB, entry *auditservice.Entry) error {
	f.entries = append(f.entries, entry)
	return f.err
}

func apiKeySetting(id, name string) *entity.Setting {
	return &entity.Setting{
		ID: id, Name: name, Type: base.SettingTypeAPIKey,
		ObjectID: "usr_1", Status: base.SettingStatusActive,
	}
}

// Taking the capability away has to take the keys with it - otherwise the user
// keeps working credentials for up to a year after being told they may not have
// any.
func TestRevokeUserAPIKeys(t *testing.T) {
	audit := &fakeAuditService{}
	uc := &UC{
		settingRepo: &fakeSettingRepo{settings: []*entity.Setting{
			apiKeySetting("set_1", "ci"), apiKeySetting("set_2", "laptop"),
		}},
		auditService: audit,
	}

	target := &entity.User{ID: "usr_1", Username: "member"}
	persistingData := &userservice.PersistingUserData{
		DeletingAccesses: []*base.PermissionResource{
			permResource(base.ResourceTypeCapability, string(base.ResourceCapAPIKeyCreate)),
		},
	}

	err := uc.cascadeRevokedCapabilities(context.Background(), nil, &basedto.Auth{},
		target, persistingData)
	if err != nil {
		t.Fatal(err)
	}

	if len(persistingData.UpsertingSettings) != 2 {
		t.Fatalf("both keys must be queued for writing back, got %d",
			len(persistingData.UpsertingSettings))
	}
	for _, setting := range persistingData.UpsertingSettings {
		if setting.DeletedAt.IsZero() {
			t.Errorf("key %q was not deleted", setting.ID)
		}
	}
	if len(audit.entries) != 2 {
		t.Fatalf("each revoked key needs its own record, got %d", len(audit.entries))
	}
	entry := audit.entries[0]
	if entry.Type != base.AuditLogTypeAPIKeyRevoke {
		t.Errorf("Type = %q", entry.Type)
	}
	// The actor is whoever ran the update; without the owner in Detail the entry
	// does not say whose key went.
	if !strings.Contains(entry.Detail, "usr_1") || !strings.Contains(entry.Detail, "member") {
		t.Errorf("the record must name the key's owner, got %q", entry.Detail)
	}
}

// Keeping the capability must leave the keys alone.
func TestCascadeLeavesKeysWhenCapabilityStays(t *testing.T) {
	audit := &fakeAuditService{}
	uc := &UC{
		settingRepo:  &fakeSettingRepo{settings: []*entity.Setting{apiKeySetting("set_1", "ci")}},
		auditService: audit,
	}

	apiKeyCap := string(base.ResourceCapAPIKeyCreate)
	persistingData := &userservice.PersistingUserData{
		UpsertingAccesses: []*entity.ACLPermission{aclRow(base.ResourceTypeCapability, apiKeyCap)},
		DeletingAccesses:  []*base.PermissionResource{permResource(base.ResourceTypeCapability, apiKeyCap)},
	}

	err := uc.cascadeRevokedCapabilities(context.Background(), nil, &basedto.Auth{},
		&entity.User{ID: "usr_1"}, persistingData)
	if err != nil {
		t.Fatal(err)
	}
	if len(persistingData.UpsertingSettings) != 0 || len(audit.entries) != 0 {
		t.Error("a capability that is kept must not touch any key")
	}
}

// Revoking a different capability must not reach the keys either.
func TestCascadeIgnoresUnrelatedCapabilities(t *testing.T) {
	audit := &fakeAuditService{}
	uc := &UC{
		settingRepo:  &fakeSettingRepo{settings: []*entity.Setting{apiKeySetting("set_1", "ci")}},
		auditService: audit,
	}

	persistingData := &userservice.PersistingUserData{
		DeletingAccesses: []*base.PermissionResource{
			permResource(base.ResourceTypeCapability, string(base.ResourceCapSecretReveal)),
		},
	}

	err := uc.cascadeRevokedCapabilities(context.Background(), nil, &basedto.Auth{},
		&entity.User{ID: "usr_1"}, persistingData)
	if err != nil {
		t.Fatal(err)
	}
	if len(persistingData.UpsertingSettings) != 0 {
		t.Error("revoking secret-reveal must not delete API keys")
	}
}

// The record and the deletion are one change: if the record cannot be written the
// transaction must fail rather than take the keys away silently.
func TestRevokeUserAPIKeysFailsClosedWhenNotRecordable(t *testing.T) {
	uc := &UC{
		settingRepo:  &fakeSettingRepo{settings: []*entity.Setting{apiKeySetting("set_1", "ci")}},
		auditService: &fakeAuditService{err: errors.New("database is down")},
	}

	persistingData := &userservice.PersistingUserData{
		DeletingAccesses: []*base.PermissionResource{
			permResource(base.ResourceTypeCapability, string(base.ResourceCapAPIKeyCreate)),
		},
	}

	err := uc.cascadeRevokedCapabilities(context.Background(), nil, &basedto.Auth{},
		&entity.User{ID: "usr_1"}, persistingData)
	if err == nil {
		t.Fatal("a revocation that cannot be recorded must fail the update")
	}
}
