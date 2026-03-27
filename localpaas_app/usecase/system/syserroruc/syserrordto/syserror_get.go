package syserrordto

import (
	"time"

	vld "github.com/tiendc/go-validator"

	"github.com/localpaas/localpaas/localpaas_app/apperrors"
	"github.com/localpaas/localpaas/localpaas_app/basedto"
	"github.com/localpaas/localpaas/localpaas_app/entity"
	"github.com/localpaas/localpaas/localpaas_app/pkg/copier"
)

type GetSysErrorReq struct {
	ID string `json:"-"`
}

func NewGetSysErrorReq() *GetSysErrorReq {
	return &GetSysErrorReq{}
}

func (req *GetSysErrorReq) Validate() apperrors.ValidationErrors {
	var validators []vld.Validator
	validators = append(validators, basedto.ValidateID(&req.ID, true, "id")...)
	return apperrors.NewValidationErrors(vld.Validate(validators...))
}

type GetSysErrorResp struct {
	Meta *basedto.Meta `json:"meta"`
	Data *SysErrorResp `json:"data"`
}

type SysErrorResp struct {
	ID         string `json:"id"`
	RequestID  string `json:"requestID"`
	Status     int    `json:"status"`
	Code       string `json:"code"`
	Detail     string `json:"detail"`
	Cause      string `json:"cause"`
	DebugLog   string `json:"debugLog"`
	StackTrace string `json:"stackTrace"`

	CreatedAt time.Time `json:"createdAt"`
}

func TransformSysError(appErr *entity.SysError) (resp *SysErrorResp, err error) {
	if err = copier.Copy(&resp, &appErr); err != nil {
		return nil, apperrors.Wrap(err)
	}
	return resp, nil
}
