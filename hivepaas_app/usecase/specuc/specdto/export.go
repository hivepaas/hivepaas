package specdto

import (
	vld "github.com/tiendc/go-validator"

	"github.com/hivepaas/hivepaas/hivepaas_app/basedto"
	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/specservice/specmodel"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/settings"
)

const (
	passphraseMaxLen = 64
)

// ExportSpecReq asks for a configuration bundle at one scope.
type ExportSpecReq struct {
	Scope *entity.ObjectScope `json:"-"`

	// SecretsMode is omit, encrypted or plaintext. It defaults to omit, the one
	// mode that needs no capability and leaks nothing.
	SecretsMode specmodel.SecretsMode `json:"secretsMode"`

	// Passphrase is required when SecretsMode is encrypted, and arrives in the
	// request body rather than the query string.
	//
	// That is why these endpoints are POST rather than GET. The access log
	// records the full path including the query - verified against a running
	// instance, which logged `path=/_/spec/export?secretsMode=omit` - and this
	// installation's own logging subsystem ships those lines to a searchable
	// store along with Traefik's access logs. A passphrase in the URL would be
	// retained there in plain text. Being a POST also keeps the response, which
	// may contain every secret in the scope, out of any intermediary cache.
	Passphrase string `json:"passphrase"`
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
	validators := make([]vld.Validator, 0, 5) //nolint:mnd
	validators = append(validators, basedto.ValidateStrIn(&req.SecretsMode, true,
		specmodel.AllSecretsModes, "secretsMode")...)
	validators = append(validators, basedto.ValidateStr(&req.Passphrase, false,
		1, passphraseMaxLen, "passphrase")...)
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
