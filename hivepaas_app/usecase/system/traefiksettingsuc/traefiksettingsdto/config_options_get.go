package traefiksettingsdto

import (
	"strings"

	"github.com/moby/moby/api/types/swarm"

	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/basedto"
	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
	"github.com/hivepaas/hivepaas/services/traefik/traefikhelper"
)

type GetConfigOptionsReq struct {
}

func NewGetConfigOptionsReq() *GetConfigOptionsReq {
	return &GetConfigOptionsReq{}
}

func (req *GetConfigOptionsReq) Validate() hperrors.ValidationErrors {
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
	OpenPorts []string `json:"openPorts,omitempty"`
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
		OpenPorts: make([]string, 0, 10), //nolint:mnd
		Args:      make([]string, 0, 20), //nolint:mnd
	}
	if traefikSvc == nil || traefikSvc.Spec.TaskTemplate.ContainerSpec == nil {
		return resp
	}

	var log bool
	for _, arg := range traefikSvc.Spec.TaskTemplate.ContainerSpec.Args {
		key, val, valid := traefikhelper.ParseCommandArg(arg)
		if !valid {
			continue
		}

		if openPort, ok := parseOpenPortFromArg(key, val); ok {
			resp.OpenPorts = append(resp.OpenPorts, openPort)
			continue
		}

		if !base.IsTraefikCmdArgSettable(key) {
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

func parseOpenPortFromArg(key, val string) (string, bool) {
	if !strings.HasSuffix(key, ".address") {
		return "", false
	}
	isTCP := strings.HasPrefix(key, "entrypoints.tcp-svc-")
	isUDP := strings.HasPrefix(key, "entrypoints.udp-svc-")
	if !isTCP && !isUDP {
		return "", false
	}

	val = strings.TrimPrefix(strings.TrimSpace(val), ":")
	if val == "" {
		return "", false
	}
	return val, true
}

func isBoolTrue(val string) bool {
	return val == "" || strings.EqualFold(val, "true") || val == "1"
}
