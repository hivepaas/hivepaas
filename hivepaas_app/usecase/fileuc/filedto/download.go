package filedto

import (
	"io"

	vld "github.com/tiendc/go-validator"

	"github.com/hivepaas/hivepaas/hivepaas_app/basedto"
	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/timeutil"
)

type DownloadFileReq struct {
	ID                      string            `json:"-" mapstructure:"-"`
	Token                   string            `json:"-" mapstructure:"token"`
	ViewInline              bool              `json:"-" mapstructure:"viewInline"`
	UsePresignURLOnFileSize int64             `json:"-" mapstructure:"-"`
	PresignExpiration       timeutil.Duration `json:"-" mapstructure:"-"`
}

func NewDownloadFileReq() *DownloadFileReq {
	return &DownloadFileReq{}
}

func (req *DownloadFileReq) Validate() hperrors.ValidationErrors {
	validators := make([]vld.Validator, 0, 10) //nolint:mnd
	// TODO: add validation
	return hperrors.NewValidationErrors(vld.Validate(validators...))
}

type DownloadFileResp struct {
	Meta *basedto.Meta         `json:"meta"`
	Data *DownloadFileDataResp `json:"data"`
}

type DownloadFileDataResp struct {
	RedirectURL   string
	ContentType   string
	ContentLength int64
	ExtraHeaders  map[string]string
	Content       io.ReadCloser
}
