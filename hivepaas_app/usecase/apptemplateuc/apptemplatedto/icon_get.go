package apptemplatedto

import (
	vld "github.com/tiendc/go-validator"

	"github.com/hivepaas/hivepaas/hivepaas_app/basedto"
	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
)

const sha256HexLen = 64

type GetAppTemplateIconReq struct {
	SHA256 string `json:"-"`
}

func NewGetAppTemplateIconReq() *GetAppTemplateIconReq {
	return &GetAppTemplateIconReq{}
}

// Validate implements interface basedto.ReqValidator
func (req *GetAppTemplateIconReq) Validate() hperrors.ValidationErrors {
	return hperrors.NewValidationErrors(vld.Validate(
		basedto.ValidateStr(&req.SHA256, true, sha256HexLen, sha256HexLen, "sha256")...))
}

// GetAppTemplateIconResp is written as the image itself, not as JSON.
type GetAppTemplateIconResp struct {
	Content     []byte
	ContentType string
}
