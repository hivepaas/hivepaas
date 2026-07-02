package sessiondto

import (
	"time"

	"github.com/hivepaas/hivepaas/hivepaas_app/apperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/copier"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/timeutil"
)

type UserResp struct {
	ID             string                  `json:"id"`
	Username       string                  `json:"username"`
	Email          string                  `json:"email"`
	Role           base.UserRole           `json:"role"`
	Status         base.UserStatus         `json:"status"`
	FullName       string                  `json:"fullName"`
	Position       string                  `json:"position"`
	Photo          string                  `json:"photo"`
	Notes          string                  `json:"notes,omitempty"`
	SecurityOption base.UserSecurityOption `json:"securityOption"`
	MfaSecret      string                  `json:"mfaSecret"`
	CreatedAt      time.Time               `json:"createdAt"`
	UpdatedAt      time.Time               `json:"updatedAt"`
	AccessExpireAt *timeutil.Date          `json:"accessExpireAt" copy:",nilonzero"`
}

func TransformUser(user *entity.User) (resp *UserResp, err error) {
	if err = copier.Copy(&resp, &user); err != nil {
		return nil, apperrors.New(err)
	}
	if user.TotpSecret != "" {
		resp.MfaSecret = "********"
	}
	return resp, nil
}
