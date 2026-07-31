package traefiksettingsdto

import (
	"github.com/moby/moby/api/types/swarm"

	"github.com/hivepaas/hivepaas/hivepaas_app/apperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/basedto"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/traefikhelper"
)

type GetConfigOptionsReq struct {
}

func NewGetConfigOptionsReq() *GetConfigOptionsReq {
	return &GetConfigOptionsReq{}
}

func (req *GetConfigOptionsReq) Validate() apperrors.ValidationErrors {
	return nil
}

type GetConfigOptionsResp struct {
	Meta *basedto.Meta      `json:"meta"`
	Data *ConfigOptionsResp `json:"data"`
}

type ConfigOptionsResp struct {
	CommandArgs []string `json:"commandArgs"`
}

func TransformConfigOptions(
	traefikSvc *swarm.Service,
) (resp *ConfigOptionsResp, err error) {
	resp = &ConfigOptionsResp{
		CommandArgs: make([]string, 0),
	}
	if traefikSvc == nil || traefikSvc.Spec.TaskTemplate.ContainerSpec == nil {
		return resp, nil
	}

	for _, arg := range traefikSvc.Spec.TaskTemplate.ContainerSpec.Args {
		key, _, valid := traefikhelper.ParseCommandArg(arg)
		if valid && base.IsTraefikCmdArgSettable(key) {
			resp.CommandArgs = append(resp.CommandArgs, arg)
		}
	}
	return resp, nil
}
