package filedto

import (
	vld "github.com/tiendc/go-validator"

	"github.com/hivepaas/hivepaas/hivepaas_app/apperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/basedto"
)

type DeleteFileReq struct {
	ID                string `json:"-" mapstructure:"-"`
	ObjectID          string `json:"-" mapstructure:"-"`
	DeletePermanently bool   `json:"-" mapstructure:"deletePermanently"`
}

func NewDeleteFileReq() *DeleteFileReq {
	return &DeleteFileReq{}
}

// Validate implements interface basedto.ReqValidator
func (req *DeleteFileReq) Validate() apperrors.ValidationErrors {
	var validators []vld.Validator
	validators = append(validators, basedto.ValidateID(&req.ID, true, "id")...)
	validators = append(validators, basedto.ValidateID(&req.ObjectID, false, "objectId")...)
	return apperrors.NewValidationErrors(vld.Validate(validators...))
}

type DeleteFileResp struct {
	Meta *basedto.Meta `json:"meta"`
}
