package basedto

import (
	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
)

type UserBaseResp struct {
	ID       string        `json:"id"`
	Username string        `json:"username"`
	Email    string        `json:"email"`
	FullName string        `json:"fullName"`
	Photo    string        `json:"photo"`
	Role     base.UserRole `json:"role"`
}

func TransformUserBase(user *entity.User) *UserBaseResp {
	if user == nil {
		return nil
	}
	return &UserBaseResp{
		ID:       user.ID,
		Username: user.Username,
		Email:    user.Email,
		FullName: user.FullName,
		Photo:    user.Photo,
		Role:     user.Role,
	}
}

func TransformUsersBase(users []*entity.User) []*UserBaseResp {
	resp, _ := TransformObjectSlice(users, func(user *entity.User) (*UserBaseResp, error) {
		return TransformUserBase(user), nil
	})
	return resp
}

func NewMissingUserResp(id string) *UserBaseResp {
	//nolint:goconst
	return &UserBaseResp{
		ID:       id,
		Username: "<missing>",
		Email:    "<missing>",
		FullName: "<missing>",
		Role:     base.UserRoleMember,
	}
}
