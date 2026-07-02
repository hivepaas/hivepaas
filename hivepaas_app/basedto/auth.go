package basedto

import (
	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/jwtsession"
)

type User struct {
	*entity.User
	AuthClaims *jwtsession.AuthClaims
}

func (u *User) Entity() *entity.User {
	if u != nil {
		return u.User
	}
	return nil
}

type Auth struct {
	User *User

	// AllowObjectIDs object IDs which the current user can access
	AllowObjectIDs []string
}
