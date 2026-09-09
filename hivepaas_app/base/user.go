package base

import "slices"

type UserRole string

const (
	UserRoleAdmin  UserRole = "admin"
	UserRoleMember UserRole = "member"
)

var (
	AllUserRoles = []UserRole{UserRoleAdmin, UserRoleMember}
)

type UserStatus string

const (
	UserStatusActive   UserStatus = "active"
	UserStatusPending  UserStatus = "pending"
	UserStatusDisabled UserStatus = "disabled"
)

var (
	AllUserStatuses = []UserStatus{UserStatusActive, UserStatusPending, UserStatusDisabled}
)

type UserSecurityOption string

const (
	UserSecurityEnforceSSO   UserSecurityOption = "enforce-sso"
	UserSecurityPassword2FA  UserSecurityOption = "password-2fa"
	UserSecurityPasswordOnly UserSecurityOption = "password-only"
)

var (
	AllUserSecurityOptions = []UserSecurityOption{UserSecurityEnforceSSO, UserSecurityPassword2FA,
		UserSecurityPasswordOnly}

	// AdminSecurityOptions are the ways an admin account may authenticate.
	//
	// An allow-list rather than a ban on password-only, because of which mistake
	// is worse. A strong option added later and forgotten here is refused for
	// admins until somebody adds it - visible the first time it is tried, and one
	// line to fix. The other way round, a weak option added later would be
	// silently accepted for the accounts that can do the most damage.
	AdminSecurityOptions = []UserSecurityOption{UserSecurityEnforceSSO, UserSecurityPassword2FA}
)

// SecurityOptionAllowedForRole reports whether the option is strong enough for
// the role.
//
// Only admins are held to it. An admin can grant themselves every module, every
// project and every capability, so a password on its own is the whole of the
// defense around all of it - and a password is the one factor that leaks without
// anybody noticing.
func SecurityOptionAllowedForRole(role UserRole, option UserSecurityOption) bool {
	if role != UserRoleAdmin {
		return true
	}
	return slices.Contains(AdminSecurityOptions, option)
}
