package filedto

import (
	vld "github.com/tiendc/go-validator"

	"github.com/hivepaas/hivepaas/hivepaas_app/basedto"
	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/settings/commandtemplateuc/commandtemplatedto"
)

const maxPassphraseLen = 1000

// LoadDataFileReq feeds an app's data file to a command run in the app, on its
// stdin: a dump a job saved, loaded back into the app's database, say.
type LoadDataFileReq struct {
	ID        string `json:"-" mapstructure:"-"`
	ProjectID string `json:"-" mapstructure:"-"`
	AppID     string `json:"-" mapstructure:"-"`

	Command *commandtemplatedto.CommandTemplateBaseReq `json:"command"`
	// Passphrase decrypts a file its job saved encrypted, named .age.
	Passphrase string `json:"passphrase"`
}

func NewLoadDataFileReq() *LoadDataFileReq {
	return &LoadDataFileReq{}
}

// ModifyRequest implements interface basedto.ReqModifier
func (req *LoadDataFileReq) ModifyRequest() error {
	if req.Command == nil {
		return nil
	}
	req.Command.Name = "-"
	req.Command.Kind = ""
	// It reads the file on its stdin, which a TTY would not pass through as is.
	req.Command.TTY = false
	return hperrors.Wrap(req.Command.ModifyRequest())
}

// Validate implements interface basedto.ReqValidator
func (req *LoadDataFileReq) Validate() hperrors.ValidationErrors {
	validators := make([]vld.Validator, 0, 5) //nolint:mnd
	validators = append(validators, basedto.ValidateID(&req.ID, true, "id")...)
	validators = append(validators, basedto.ValidateCond(req.Command != nil, "command")...)
	if req.Command != nil {
		validators = append(validators, req.Command.Validate("command")...)
	}
	validators = append(validators, basedto.ValidateStr(&req.Passphrase, false, 1, maxPassphraseLen,
		"passphrase")...)
	validators = append(validators, basedto.ValidatePlainSecret(&req.Passphrase, "passphrase")...)
	validators = append(validators, basedto.ValidateNotMaskedSecret(&req.Passphrase, "passphrase")...)
	return hperrors.NewValidationErrors(vld.Validate(validators...))
}

type LoadDataFileResp struct {
	Meta *basedto.Meta         `json:"meta"`
	Data *LoadDataFileDataResp `json:"data"`
}

type LoadDataFileDataResp struct {
	// Task is the load's.
	Task *basedto.ObjectIDResp `json:"task"`
}
