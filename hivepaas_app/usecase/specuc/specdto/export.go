package specdto

import (
	vld "github.com/tiendc/go-validator"

	"github.com/hivepaas/hivepaas/hivepaas_app/basedto"
	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/specservice/specmodel"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/settings"
)

// ExportSpecReq asks for a configuration bundle at one scope.
type ExportSpecReq struct {
	ProjectID    string `json:"-"`
	ProjectEnvID string `json:"-"`
	AppID        string `json:"-"`

	// SecretsMode is omit, encrypted or plaintext. It defaults to omit, the one
	// mode that needs no capability and leaks nothing.
	SecretsMode specmodel.SecretsMode `json:"-" mapstructure:"secretsMode"`
	// Passphrase is required when SecretsMode is encrypted.
	Passphrase string `json:"-" mapstructure:"passphrase"`
}

func NewExportSpecReq() *ExportSpecReq {
	return &ExportSpecReq{SecretsMode: specmodel.SecretsModeOmit}
}

// ModifyRequest defaults the mode, so a caller who omits it gets the harmless
// one rather than an error.
func (req *ExportSpecReq) ModifyRequest() {
	if req.SecretsMode == "" {
		req.SecretsMode = specmodel.SecretsModeOmit
	}
}

func (req *ExportSpecReq) Validate() hperrors.ValidationErrors {
	validators := make([]vld.Validator, 0, 4) //nolint:mnd
	if req.ProjectID != "" {
		validators = append(validators, basedto.ValidateID(&req.ProjectID, true, "projectId")...)
	}
	if req.AppID != "" {
		validators = append(validators, basedto.ValidateID(&req.AppID, true, "appId")...)
	}
	return hperrors.NewValidationErrors(vld.Validate(validators...))
}

// ExportSpecResp carries the bundle itself. The body is the archive, so the
// report travels in a header - see spechandler.
type ExportSpecResp struct {
	Data *settings.BaseDownloadDataResp
	// Summary is what the response header carries. The full report is inside
	// the bundle as report.yaml, because a header cannot hold it.
	Summary *specmodel.ReportSummary
}
