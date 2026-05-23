package basedto

import (
	"github.com/localpaas/localpaas/localpaas_app/base"
	"github.com/localpaas/localpaas/localpaas_app/entity"
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
