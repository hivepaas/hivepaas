package filedto

import (
	vld "github.com/tiendc/go-validator"

	"github.com/hivepaas/hivepaas/hivepaas_app/apperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/basedto"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/timeutil"
)

type GetFileDownloadURLReq struct {
	ID           string            `json:"-" mapstructure:"-"`
	ObjectID     string            `json:"-" mapstructure:"-"`
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
	validators = append(validators, basedto.ValidateID(&req.ID, true, "id")...)
	validators = append(validators, basedto.ValidateID(&req.ObjectID, false, "objectId")...)
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
