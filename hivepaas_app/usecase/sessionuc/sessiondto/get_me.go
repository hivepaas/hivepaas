package sessiondto

import (
	"github.com/hivepaas/hivepaas/hivepaas_app/basedto"
	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
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
	// Timezone is the installation's, a zone name such as America/New_York:
	// what a schedule's hours are read in.
	Timezone string `json:"timezone"`
	// Server is what the installation runs, for a client to know what it may
	// ask of it - the CLI reads it on login.
	Server *ServerInfoResp `json:"server"`
}

// ServerInfoResp is the release an installation runs and the API it answers.
type ServerInfoResp struct {
	// Version is the release, such as v1.0.0-beta4.
	Version string `json:"version"`
	// APILevel is the level of the API: raised whenever the request of a write
	// operation that already existed changes.
	APILevel int `json:"apiLevel"`
	// MinCLIAPILevel is the lowest API level a HivePaaS CLI may be built for to
	// write: one below it is answered 426, as it would write objects back
	// without the fields it does not know. It may read.
	MinCLIAPILevel int `json:"minCliApiLevel"`
}

func TransformUserDetails(user *entity.User) (resp *userdto.UserDetailsResp, err error) {
	resp, err = userdto.TransformUserDetails(user)
	if err != nil {
		return nil, hperrors.Wrap(err)
	}
	return resp, nil
}
