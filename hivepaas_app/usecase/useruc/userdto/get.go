package userdto

import (
	"slices"
	"strings"
	"time"

	vld "github.com/tiendc/go-validator"

	"github.com/hivepaas/hivepaas/hivepaas_app/apperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/basedto"
	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
)

type GetUserReq struct {
	ID          string `json:"-"`
	GetAccesses bool   `json:"-" mapstructure:"getAccesses"`
}

func NewGetUserReq() *GetUserReq {
	return &GetUserReq{}
}

func (req *GetUserReq) Validate() apperrors.ValidationErrors {
	return apperrors.NewValidationErrors(vld.Validate(basedto.ValidateID(&req.ID, true, "id")...))
}

type GetUserResp struct {
	Meta *basedto.Meta    `json:"meta"`
	Data *UserDetailsResp `json:"data"`
}

type UserDetailsResp struct {
	*UserResp
	ProjectAccesses basedto.ObjectAccessSliceResp `json:"projectAccesses"`
	ModuleAccesses  basedto.ObjectAccessSliceResp `json:"moduleAccesses"`
}

type UserResp struct {
	ID               string                  `json:"id"`
	Username         string                  `json:"username"`
	Email            string                  `json:"email"`
	Role             base.UserRole           `json:"role"`
	Status           base.UserStatus         `json:"status"`
	FullName         string                  `json:"fullName"`
	Photo            string                  `json:"photo"`
	Position         string                  `json:"position"`
	SecurityOption   base.UserSecurityOption `json:"securityOption"`
	MfaTotpActivated bool                    `json:"mfaTotpActivated,omitempty"`
	Notes            string                  `json:"notes,omitempty"`

	CreatedAt      time.Time  `json:"createdAt"`
	UpdatedAt      time.Time  `json:"updatedAt"`
	AccessExpireAt *time.Time `json:"accessExpireAt" copy:",nilonzero"`
	LastAccess     *time.Time `json:"lastAccess" copy:",nilonzero"`
}

func TransformUserDetails(user *entity.User) (resp *UserDetailsResp, err error) {
	userResp, err := TransformUser(user)
	if err != nil {
		return nil, apperrors.New(err)
	}
	resp = &UserDetailsResp{
		UserResp: userResp,
	}

	for _, access := range user.Accesses {
		if access.ResourceType == base.ResourceTypeProject && access.ResourceProject != nil {
			resp.ProjectAccesses = append(resp.ProjectAccesses, &basedto.ObjectAccessResp{
				NamedObjectResp: basedto.NamedObjectResp{
					ID:   access.ResourceProject.ID,
					Name: access.ResourceProject.Name,
				},
				Access: access.Actions,
			})
			continue
		}
		if access.ResourceType == base.ResourceTypeModule {
			resp.ModuleAccesses = append(resp.ModuleAccesses, &basedto.ObjectAccessResp{
				NamedObjectResp: basedto.NamedObjectResp{
					ID: access.ResourceID,
				},
				Access: access.Actions,
			})
			continue
		}
	}

	// Sort project accesses by project names
	slices.SortStableFunc(resp.ProjectAccesses, func(a, b *basedto.ObjectAccessResp) int {
		return strings.Compare(a.Name, b.Name)
	})

	return resp, nil
}
