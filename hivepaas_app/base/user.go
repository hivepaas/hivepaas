package base

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
)
