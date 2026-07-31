package traefiksettingsdto

import (
	"strings"

	vld "github.com/tiendc/go-validator"

	"github.com/hivepaas/hivepaas/hivepaas_app/apperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/basedto"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/traefikhelper"
)

type UpdateConfigOptionsReq struct {
	CommandArgs []string `json:"commandArgs"`

	ParsedCommandArgs [][]string `json:"-"` // Use internally only
}

func NewUpdateConfigOptionsReq() *UpdateConfigOptionsReq {
	return &UpdateConfigOptionsReq{}
}

func (req *UpdateConfigOptionsReq) ModifyRequest() error {
	req.ParsedCommandArgs = make([][]string, 0, len(req.CommandArgs))
	for _, arg := range req.CommandArgs {
		arg = strings.TrimSpace(arg)
		key, val, valid := traefikhelper.ParseCommandArg(arg)
		if valid {
			req.ParsedCommandArgs = append(req.ParsedCommandArgs, []string{key, val, arg})
		}
	}
	return nil
}

// Validate implements interface basedto.ReqValidator
func (req *UpdateConfigOptionsReq) Validate() apperrors.ValidationErrors {
	var validators []vld.Validator
	validators = append(validators, vld.SliceUniqueBy(req.ParsedCommandArgs, func(kv []string) string {
		return kv[0]
	}).OnError(
		vld.SetField("commandArgs", nil),
		vld.SetCustomKey("ERR_VLD_VALUES_NON_UNIQUE"),
	))
	for _, kv := range req.ParsedCommandArgs {
		key := kv[0]
		if !base.IsTraefikCmdArgSettable(key) {
			validators = append(validators, vld.Must(false).OnError(
				vld.SetField("commandArgs", nil),
				vld.SetCustomKey("ERR_VLD_VALUE_UNALLOWED"),
				vld.SetParam("Value", key),
			))
		}
	}
	return apperrors.NewValidationErrors(vld.Validate(validators...))
}

type UpdateConfigOptionsResp struct {
	Meta *basedto.Meta `json:"meta"`
}
