package useruc

import (
	"context"
	"errors"
	"testing"

	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/basedto"
	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/infra/database"
	"github.com/hivepaas/hivepaas/hivepaas_app/permission"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/timeutil"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/userservice"
)

// spyPermissionManager records whether it was asked, and lets a test stand in for
// an admin (every current row is replaceable) or refuse the change outright.
type spyPermissionManager struct {
	permission.Manager
	called      bool
	desired     []*entity.ACLPermission
	replaceable []*entity.ACLPermission
	err         error
}

func (s *spyPermissionManager) AuthorizeAccessChanges(_ context.Context, _ database.IDB,
	_ *basedto.Auth, desired, current []*entity.ACLPermission) ([]*entity.ACLPermission, error) {
	s.called = true
	s.desired = desired
	if s.err != nil {
		return nil, s.err
	}
	if s.replaceable != nil {
		return s.replaceable, nil
	}
	return current, nil // what an admin gets back
}

func capabilityRow(userID string) *entity.ACLPermission {
	return &entity.ACLPermission{
		SubjectType:  base.SubjectTypeUser,
		SubjectID:    userID,
		ResourceType: base.ResourceTypeCapability,
		ResourceID:   string(base.ResourceCapSecretReveal),
		Actions:      base.AccessActions{Exec: true},
	}
}

// A request carrying only `capabilities` used to produce an empty resource-type
// list, and authorizeAccessChanges returns early on an empty list - so the grant
// was written without anybody being asked whether the caller may hand it out.
func TestCapabilityOnlyUpdateIsAuthorized(t *testing.T) {
	spy := &spyPermissionManager{err: errors.New("refused")}
	uc := &UC{permissionManager: spy}

	persistingData := &userservice.PersistingUserData{}
	uc.preparePersistingUserCapabilities(&entity.User{ID: "usr_target"},
		basedto.CapabilitySliceReq{base.ResourceCapSecretReveal},
		timeutil.NowUTC(), persistingData)

	// Exactly what UpdateUser passes when the request carries only capabilities.
	resourceTypes := accessResourceTypesToReplace(false, true, false)

	err := uc.authorizeAccessChanges(context.Background(), nil, &basedto.Auth{},
		&entity.User{ID: "usr_target"}, resourceTypes, persistingData)

	if !spy.called {
		t.Fatal("the grant was applied without any authorization check")
	}
	if err == nil {
		t.Fatal("a refused change must fail the update")
	}
	if len(spy.desired) != 1 || spy.desired[0].ResourceType != base.ResourceTypeCapability {
		t.Errorf("the capability row must be submitted for checking, got %+v", spy.desired)
	}
}

// Dropping a capability from the list has to actually take it away.
func TestCapabilityIsRevocable(t *testing.T) {
	spy := &spyPermissionManager{}
	uc := &UC{permissionManager: spy}

	target := &entity.User{
		ID:       "usr_target",
		Accesses: []*entity.ACLPermission{capabilityRow("usr_target")},
	}

	// An empty-but-present list: "the user should hold no capabilities".
	persistingData := &userservice.PersistingUserData{}
	uc.preparePersistingUserCapabilities(target, basedto.CapabilitySliceReq{},
		timeutil.NowUTC(), persistingData)
	if len(persistingData.UpsertingAccesses) != 0 {
		t.Fatal("an empty list must write no grant")
	}

	err := uc.authorizeAccessChanges(context.Background(), nil, &basedto.Auth{},
		target, accessResourceTypesToReplace(false, true, false), persistingData)
	if err != nil {
		t.Fatal(err)
	}

	if len(persistingData.DeletingAccesses) != 1 {
		t.Fatalf("the existing capability must be queued for removal, got %d rows",
			len(persistingData.DeletingAccesses))
	}
	if got := persistingData.DeletingAccesses[0]; got.ResourceType != base.ResourceTypeCapability ||
		got.ResourceID != string(base.ResourceCapSecretReveal) {
		t.Errorf("wrong row queued for removal: %+v", got)
	}
}

// Rows the actor cannot reach must survive an update rather than be swept away by
// a wholesale replace - the same rule the other grant types follow.
func TestCapabilityOutOfReachIsLeftAlone(t *testing.T) {
	spy := &spyPermissionManager{replaceable: []*entity.ACLPermission{}}
	uc := &UC{permissionManager: spy}

	target := &entity.User{
		ID:       "usr_target",
		Accesses: []*entity.ACLPermission{capabilityRow("usr_target")},
	}

	persistingData := &userservice.PersistingUserData{}
	err := uc.authorizeAccessChanges(context.Background(), nil, &basedto.Auth{},
		target, accessResourceTypesToReplace(false, true, false), persistingData)
	if err != nil {
		t.Fatal(err)
	}
	if len(persistingData.DeletingAccesses) != 0 {
		t.Error("a row the actor cannot revoke must not be removed")
	}
}

func TestPreparePersistingUserCapabilities(t *testing.T) {
	uc := &UC{}
	persistingData := &userservice.PersistingUserData{}

	uc.preparePersistingUserCapabilities(&entity.User{ID: "usr_1"},
		basedto.CapabilitySliceReq{base.ResourceCapSecretReveal},
		timeutil.NowUTC(), persistingData)

	if len(persistingData.UpsertingAccesses) != 1 {
		t.Fatalf("expected 1 ACL row, got %d", len(persistingData.UpsertingAccesses))
	}
	row := persistingData.UpsertingAccesses[0]
	if row.ResourceType != base.ResourceTypeCapability {
		t.Errorf("ResourceType = %q", row.ResourceType)
	}
	if row.SubjectType != base.SubjectTypeUser || row.SubjectID != "usr_1" {
		t.Errorf("wrong subject: %s/%s", row.SubjectType, row.SubjectID)
	}
	// The stored action has to be the one checkCapability asks for, or the grant
	// is written and then never matches.
	if !row.Actions.Allows(base.ActionTypeExecute) {
		t.Errorf("the row must carry the execute action, got %+v", row.Actions)
	}
}
