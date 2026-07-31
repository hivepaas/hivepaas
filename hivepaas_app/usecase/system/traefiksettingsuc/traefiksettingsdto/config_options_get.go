package traefiksettingsdto

import (
	"strings"

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
	StartupCommand *StartupCommandResp `json:"startupCommand"`
}

type StartupCommandResp struct {
	LogLevel  string   `json:"logLevel,omitempty"`
	AccessLog bool     `json:"accessLog,omitempty"`
	HTTP3     bool     `json:"http3,omitempty"`
	FastProxy bool     `json:"fastProxy,omitempty"`
	Args      []string `json:"args"`
}

func TransformConfigOptions(
	traefikSvc *swarm.Service,
) (resp *ConfigOptionsResp) {
	resp = &ConfigOptionsResp{
		StartupCommand: TransformStartupCommand(traefikSvc),
	}
	return resp
}

func TransformStartupCommand(
	traefikSvc *swarm.Service,
) (resp *StartupCommandResp) {
	resp = &StartupCommandResp{
		Args: make([]string, 0, 20), //nolint:mnd
	}
	if traefikSvc == nil || traefikSvc.Spec.TaskTemplate.ContainerSpec == nil {
		return resp
	}

	var log bool
	for _, arg := range traefikSvc.Spec.TaskTemplate.ContainerSpec.Args {
		key, val, valid := traefikhelper.ParseCommandArg(arg)
		if !valid || !base.IsTraefikCmdArgSettable(key) {
			continue
		}

		switch key {
		case "log":
			log = isBoolTrue(val)
		case "log.level":
			resp.LogLevel = val
		case "accesslog":
			resp.AccessLog = isBoolTrue(val)
		case "entrypoints.websecure.http3":
			resp.HTTP3 = isBoolTrue(val)
		case "experimental.fastproxy":
			resp.FastProxy = isBoolTrue(val)
		default:
			resp.Args = append(resp.Args, arg)
		}
	}

	if resp.LogLevel == "" && log {
		resp.LogLevel = "default"
	}

	return resp
}

func isBoolTrue(val string) bool {
	return val == "" || strings.EqualFold(val, "true") || val == "1"
}
