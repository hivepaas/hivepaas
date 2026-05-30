package filedto

import (
	vld "github.com/tiendc/go-validator"

	"github.com/localpaas/localpaas/localpaas_app/apperrors"
	"github.com/localpaas/localpaas/localpaas_app/basedto"
	"github.com/localpaas/localpaas/localpaas_app/pkg/timeutil"
)

type GetFileDownloadURLReq struct {
	ID           string            `json:"-" mapstructure:"-"`
	Expiration   timeutil.Duration `json:"-" mapstructure:"-"`
	RequireLogin bool              `json:"-" mapstructure:"-"`
	ViewInline   bool              `json:"-" mapstructure:"viewInline"`
	CloudPresign bool              `json:"-" mapstructure:"-"`
}

func NewGetFileDownloadURLReq() *GetFileDownloadURLReq {
	return &GetFileDownloadURLReq{}
}

func (req *GetFileDownloadURLReq) Validate() apperrors.ValidationErrors {
	var validators []vld.Validator
	// TODO: add validation
	return apperrors.NewValidationErrors(vld.Validate(validators...))
}

type GetFileDownloadURLResp struct {
	Meta *basedto.Meta            `json:"meta"`
	Data *FileDownloadURLDataResp `json:"data"`
}

type FileDownloadURLDataResp struct {
	URL        string            `json:"url"`
	Expiration timeutil.Duration `json:"expiration"`
}
