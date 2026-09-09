package useruc

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/useruc/userdto"
)

func ptr[T any](v T) *T { return &v }

// The update request carries the role and the security option as optional
// pointers, so the rule has to be read against the result: the dangerous
// combinations arrive with either half missing.
func TestEnsureAdminSecurityOption(t *testing.T) {
	admin2FA := &entity.User{Role: base.UserRoleAdmin, SecurityOption: base.UserSecurityPassword2FA}
	memberPassword := &entity.User{Role: base.UserRoleMember, SecurityOption: base.UserSecurityPasswordOnly}
	legacyAdmin := &entity.User{Role: base.UserRoleAdmin, SecurityOption: base.UserSecurityPasswordOnly}

	tests := []struct {
		name    string
		user    *entity.User
		req     *userdto.UpdateUserReq
		refused bool
	}{
		{
			// The request carries a role and no security option at all.
			name:    "promoting a password-only member to admin",
			user:    memberPassword,
			req:     &userdto.UpdateUserReq{Role: ptr(base.UserRoleAdmin)},
			refused: true,
		},
		{
			// And here the other half is the one missing.
			name:    "weakening an admin already on 2FA",
			user:    admin2FA,
			req:     &userdto.UpdateUserReq{SecurityOption: ptr(base.UserSecurityPasswordOnly)},
			refused: true,
		},
		{
			name: "promoting to admin and strengthening in the same request",
			user: memberPassword,
			req: &userdto.UpdateUserReq{
				Role:           ptr(base.UserRoleAdmin),
				SecurityOption: ptr(base.UserSecurityPassword2FA),
			},
		},
		{
			name: "demoting an admin to member frees the option",
			user: legacyAdmin,
			req: &userdto.UpdateUserReq{
				Role:           ptr(base.UserRoleMember),
				SecurityOption: ptr(base.UserSecurityPasswordOnly),
			},
		},
		{
			// The bootstrap admin is seeded password-only, so accounts in this
			// state exist on every install. Refusing every edit to them would take
			// the disable button with it.
			name:    "disabling an account that predates the rule",
			user:    legacyAdmin,
			req:     &userdto.UpdateUserReq{Status: ptr(base.UserStatusDisabled)},
			refused: false,
		},
		{
			// The user form sends every field on every save, including the two
			// unchanged ones. That must not turn a rename into a refusal.
			name: "renaming an account that predates the rule",
			user: legacyAdmin,
			req: &userdto.UpdateUserReq{
				FullName:       "New Name",
				Role:           ptr(base.UserRoleAdmin),
				SecurityOption: ptr(base.UserSecurityPasswordOnly),
			},
			refused: false,
		},
		{
			name:    "an unrelated edit to a compliant admin",
			user:    admin2FA,
			req:     &userdto.UpdateUserReq{FullName: "New Name"},
			refused: false,
		},
		{
			name:    "a member may stay on a password alone",
			user:    memberPassword,
			req:     &userdto.UpdateUserReq{SecurityOption: ptr(base.UserSecurityPasswordOnly)},
			refused: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// A copy, because the table shares the users between cases.
			user := *tt.user
			err := ensureAdminSecurityOption(tt.req, &user)

			if tt.refused {
				assert.True(t, errors.Is(err, hperrors.ErrUserAdminSecurityOptionWeak),
					"expected the change to be refused, got %v", err)
				return
			}
			assert.NoError(t, err)
		})
	}
}
