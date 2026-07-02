package nodedto

import (
	"github.com/hivepaas/hivepaas/hivepaas_app/apperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/basedto"
)

type GetNodeJoinCommandReq struct {
	JoinAsManager bool `json:"-" mapstructure:"joinAsManager"`
}

func NewGetNodeJoinCommandReq() *GetNodeJoinCommandReq {
	return &GetNodeJoinCommandReq{}
}

// Validate implements interface basedto.ReqValidator
func (req *GetNodeJoinCommandReq) Validate() apperrors.ValidationErrors {
	return nil
}

type GetNodeJoinCommandResp struct {
	Meta *basedto.Meta               `json:"meta"`
	Data *GetNodeJoinCommandDataResp `json:"data"`
}

type GetNodeJoinCommandDataResp struct {
	Command string `json:"command"`
}
