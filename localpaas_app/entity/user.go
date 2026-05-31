package entity

import (
	"time"

	"github.com/localpaas/localpaas/localpaas_app/base"
	"github.com/localpaas/localpaas/localpaas_app/config"
	"github.com/localpaas/localpaas/localpaas_app/pkg/timeutil"
)

var (
	UserUpsertingConflictCols = []string{"id"}
	UserUpsertingUpdateCols   = []string{"username", "email", "role", "status", "full_name", "position",
		"photo", "photo_id", "notes", "security_option", "totp_secret", "password", "access_expire_at",
		"last_access", "updated_at", "deleted_at"}
	UserDefaultExcludeColumns = []string{"notes", "password"}
)

type User struct {
	ID       string          `bun:",pk" json:"id"`
	Username string          `json:"username"`
	Email    string          `bun:",nullzero" json:"email"`
	Role     base.UserRole   `json:"role"`
	Status   base.UserStatus `json:"status"`
	FullName string          `json:"fullName,omitempty"`
	Position string          `bun:",nullzero" json:"position,omitempty"`
	Photo    string          `bun:",nullzero" json:"photo,omitempty"`
	PhotoID  string          `bun:",nullzero" json:"photoId,omitempty"`
	Notes    string          `bun:",nullzero" json:"notes,omitempty"`

	SecurityOption base.UserSecurityOption `json:"securityOption"`
	TotpSecret     string                  `bun:",nullzero" json:"totpSecret,omitempty"`
	Password       string                  `bun:",nullzero" json:"password,omitempty"`

	CreatedAt      time.Time `bun:",default:current_timestamp" json:"createdAt"`
	UpdatedAt      time.Time `bun:",default:current_timestamp" json:"updatedAt"`
	AccessExpireAt time.Time `bun:",nullzero" json:"accessExpireAt,omitzero"`
	DeletedAt      time.Time `bun:",soft_delete,nullzero" json:"deletedAt,omitzero"`
	LastAccess     time.Time `bun:",nullzero" json:"lastAccess,omitzero"`

	PhotoData *BinObject       `bun:"rel:has-one,join:photo_id=id" json:"photoData,omitempty"`
	Accesses  []*ACLPermission `bun:"rel:has-many,join:id=subj_id" json:"accesses,omitempty"`
}

// GetID implements IDEntity interface
func (u *User) GetID() string {
	return u.ID
}

func (u *User) IsAccessExpired() bool {
	return !u.AccessExpireAt.IsZero() && u.AccessExpireAt.Before(timeutil.NowUTC())
}

func (u *User) IsAdmin() bool {
	return u.Role == base.UserRoleAdmin
}

func (u *User) GetSettingScope() *base.ObjectScope {
	return &base.ObjectScope{
		UserID: u.ID,
	}
}

func (u *User) IsDemoUser() bool {
	return u.ID == config.Current.Users.Demo.UserID
}
