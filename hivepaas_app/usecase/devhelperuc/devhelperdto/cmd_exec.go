package devhelperdto

import (
	"github.com/hivepaas/hivepaas/hivepaas_app/apperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/basedto"
)

type ExecuteCmdReq struct {
	Cmd      string   `json:"cmd"`
	CmdArray []string `json:"cmdArray"`
	Dir      string   `json:"dir"`
	Env      []string `json:"env"`
}

func NewExecuteCmdReq() *ExecuteCmdReq {
	return &ExecuteCmdReq{}
}

func (req *ExecuteCmdReq) Validate() apperrors.ValidationErrors {
	return nil
}

type ExecuteCmdResp struct {
	Meta *basedto.Meta       `json:"meta"`
	Data *ExecuteCmdDataResp `json:"data"`
}

type ExecuteCmdDataResp struct {
	Error    string   `json:"error"`
	Output   []string `json:"output"`
	ExitCode int      `json:"exitCode"`
}
