package userdto

import (
	"testing"

	vld "github.com/tiendc/go-validator"

	"github.com/hivepaas/hivepaas/hivepaas_app/base"
)

// An admin can grant themselves every module, every project and every
// capability, so a password on its own is the whole of the defense around all of
// it - and a password is the one factor that leaks without anybody noticing.
func TestValidateAdminSecurityOption(t *testing.T) {
	tests := []struct {
		name    string
		role    base.UserRole
		option  base.UserSecurityOption
		wantErr bool
	}{
		{"admin with password only", base.UserRoleAdmin, base.UserSecurityPasswordOnly, true},
		{"admin with 2FA", base.UserRoleAdmin, base.UserSecurityPassword2FA, false},
		{"admin with enforced SSO", base.UserRoleAdmin, base.UserSecurityEnforceSSO, false},
		// A member cannot grant themselves anything, so the choice stays theirs.
		{"member with password only", base.UserRoleMember, base.UserSecurityPasswordOnly, false},
		{"member with 2FA", base.UserRoleMember, base.UserSecurityPassword2FA, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := vld.Validate(validateAdminSecurityOption(tt.role, tt.option, "securityOption")...)
			if tt.wantErr && len(err) == 0 {
				t.Error("expected a validation error, got none")
			}
			if !tt.wantErr && len(err) > 0 {
				t.Errorf("expected no validation error, got %v", err)
			}
		})
	}
}

// The rule is worth nothing if the invite request does not run it. Both requests
// below are otherwise complete, so the only difference between them is the one
// under test.
func TestInviteUserReqRunsTheAdminSecurityRule(t *testing.T) {
	newReq := func(option base.UserSecurityOption) *InviteUserReq {
		return &InviteUserReq{
			Email:          "someone@example.com",
			Role:           base.UserRoleAdmin,
			SecurityOption: option,
		}
	}

	if errs := newReq(base.UserSecurityPassword2FA).Validate(); len(errs) > 0 {
		t.Fatalf("an admin invited with 2FA must be accepted, got %v", errs)
	}
	if errs := newReq(base.UserSecurityPasswordOnly).Validate(); len(errs) == 0 {
		t.Error("an admin invited with a password and nothing else must be refused")
	}
}

// Every security option has to be a deliberate decision for admins. One added to
// AllUserSecurityOptions but not to AdminSecurityOptions is refused for them,
// which is the safe direction - but it should be somebody's choice rather than a
// surprise, so this is the line that makes them look.
func TestAdminSecurityOptionsAreADeliberateSubset(t *testing.T) {
	for _, option := range base.AdminSecurityOptions {
		if !base.SecurityOptionAllowedForRole(base.UserRoleAdmin, option) {
			t.Errorf("%s is listed for admins but refused for them", option)
		}
	}
	if base.SecurityOptionAllowedForRole(base.UserRoleAdmin, base.UserSecurityPasswordOnly) {
		t.Error("password-only must never be enough for an admin")
	}
	if len(base.AdminSecurityOptions) != 2 {
		t.Errorf("a security option was added or removed - decide whether an admin may use it, got %v",
			base.AdminSecurityOptions)
	}
}
