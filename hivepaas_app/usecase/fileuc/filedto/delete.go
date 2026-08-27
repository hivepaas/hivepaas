package filedto

import (
	vld "github.com/tiendc/go-validator"

	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/basedto"
	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
)

type DeleteFileReq struct {
	ID                       string                `json:"-" mapstructure:"-"`
	Scope                    *base.ObjectScopeType `json:"-" mapstructure:"-"`
	ObjectID                 string                `json:"-" mapstructure:"-"`
	Types                    []base.FileType       `json:"-" mapstructure:"-"`
	Kinds                    []base.FileKind       `json:"-"`
	DeletePermanentlyIfLocal bool                  `json:"-" mapstructure:"deletePermanentlyIfLocal"`
	DeletePermanently        bool                  `json:"-" mapstructure:"deletePermanently"`
}

func NewDeleteFileReq() *DeleteFileReq {
	return &DeleteFileReq{}
}

// Validate implements interface basedto.ReqValidator
func (req *DeleteFileReq) Validate() hperrors.ValidationErrors {
	validators := make([]vld.Validator, 0, 10) //nolint:mnd
	validators = append(validators, basedto.ValidateID(&req.ID, true, "id")...)
	validators = append(validators, basedto.ValidateID(&req.ObjectID, false, "objectId")...)
	return hperrors.NewValidationErrors(vld.Validate(validators...))
}

type DeleteFileResp struct {
	Meta *basedto.Meta `json:"meta"`
}
