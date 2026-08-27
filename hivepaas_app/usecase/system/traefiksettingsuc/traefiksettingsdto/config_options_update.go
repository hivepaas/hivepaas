package traefiksettingsdto

import (
	"strings"

	vld "github.com/tiendc/go-validator"

	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/basedto"
	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
	"github.com/hivepaas/hivepaas/services/traefik/traefikhelper"
)

type UpdateConfigOptionsReq struct {
	StartupCommand *StartupCommandReq `json:"startupCommand"`
}

type StartupCommandReq struct {
	LogLevel  string   `json:"logLevel"`
	AccessLog bool     `json:"accessLog"`
	HTTP3     bool     `json:"http3"`
	FastProxy bool     `json:"fastProxy"`
	Args      []string `json:"args"`

	ParsedArgs [][]string `json:"-"` // Use internally only
}

func (req *StartupCommandReq) modifyRequest() error {
	if req == nil {
		return nil
	}
	req.LogLevel = strings.TrimSpace(req.LogLevel)
	req.ParsedArgs = make([][]string, 0, len(req.Args))
	for _, arg := range req.Args {
		arg = strings.TrimSpace(arg)
		key, val, valid := traefikhelper.ParseCommandArg(arg)
		if valid {
			req.ParsedArgs = append(req.ParsedArgs, []string{key, val, arg})
		}
	}
	return nil
}

func (req *StartupCommandReq) validate(field string) (res []vld.Validator) {
	if req == nil {
		return
	}
	if field != "" {
		field += "."
	}
	res = append(res, vld.SliceUniqueBy(req.ParsedArgs, func(kv []string) string {
		return kv[0]
	}).OnError(
		vld.SetField(field+"args", nil),
		vld.SetCustomKey("ERR_VLD_VALUES_NON_UNIQUE"),
	))
	for _, kv := range req.ParsedArgs {
		key := kv[0]
		if !base.IsTraefikCmdArgSettable(key) {
			res = append(res, vld.Must(false).OnError(
				vld.SetField(field+"args", nil),
				vld.SetCustomKey("ERR_VLD_VALUE_UNALLOWED"),
				vld.SetParam("Value", key),
			))
		}
	}
	return res
}

func NewUpdateConfigOptionsReq() *UpdateConfigOptionsReq {
	return &UpdateConfigOptionsReq{}
}

func (req *UpdateConfigOptionsReq) ModifyRequest() error {
	return req.StartupCommand.modifyRequest()
}

// Validate implements interface basedto.ReqValidator
func (req *UpdateConfigOptionsReq) Validate() hperrors.ValidationErrors {
	validators := make([]vld.Validator, 0, 10) //nolint:mnd
	validators = append(validators, req.StartupCommand.validate("startupCommand")...)
	return hperrors.NewValidationErrors(vld.Validate(validators...))
}

type UpdateConfigOptionsResp struct {
	Meta *basedto.Meta `json:"meta"`
}
