package sessiondto

import (
	"github.com/hivepaas/hivepaas/hivepaas_app/basedto"
	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/system/getstarteduc/getstarteddto"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/useruc/userdto"
)

type GetMeReq struct {
	GetAccesses bool `json:"-" mapstructure:"getAccesses"`
}

func NewGetMeReq() *GetMeReq {
	return &GetMeReq{}
}

type GetMeResp struct {
	Meta *basedto.Meta  `json:"meta"`
	Data *GetMeDataResp `json:"data"`
}

type GetMeDataResp struct {
	NextStep string                   `json:"nextStep,omitempty"`
	User     *userdto.UserDetailsResp `json:"user"`
	// SetupChecklist is what the installation still has to do, for an admin
	// while nextStep is hivepaas/get-started.
	SetupChecklist *getstarteddto.ChecklistResp `json:"setupChecklist,omitempty"`
}

func TransformUserDetails(user *entity.User) (resp *userdto.UserDetailsResp, err error) {
	resp, err = userdto.TransformUserDetails(user)
	if err != nil {
		return nil, hperrors.Wrap(err)
	}
	return resp, nil
}
