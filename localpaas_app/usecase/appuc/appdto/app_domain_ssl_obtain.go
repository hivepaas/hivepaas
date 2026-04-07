package appdto

import (
	"strings"

	vld "github.com/tiendc/go-validator"
	"github.com/tiendc/gofn"

	"github.com/localpaas/localpaas/localpaas_app/apperrors"
	"github.com/localpaas/localpaas/localpaas_app/base"
	"github.com/localpaas/localpaas/localpaas_app/basedto"
)

type ObtainDomainSSLReq struct {
	ProjectID string          `json:"-"`
	AppID     string          `json:"-"`
	Domain    string          `json:"domain"`
	Email     string          `json:"email"`
	KeyType   base.SSLKeyType `json:"keyType"`
}

func NewObtainDomainSSLReq() *ObtainDomainSSLReq {
	return &ObtainDomainSSLReq{}
}

func (req *ObtainDomainSSLReq) ModifyRequest() error {
	req.Domain = strings.TrimSpace(strings.ToLower(req.Domain))
	req.KeyType = gofn.Coalesce(req.KeyType, base.SSLKeyTypeDefault)
	return nil
}

// Validate implements interface basedto.ReqValidator
func (req *ObtainDomainSSLReq) Validate() apperrors.ValidationErrors {
	var validators []vld.Validator
	validators = append(validators, basedto.ValidateID(&req.ProjectID, true, "projectID")...)
	validators = append(validators, basedto.ValidateID(&req.AppID, true, "appID")...)
	validators = append(validators, basedto.ValidateStr(&req.Domain, true, 1,
		base.DomainNameMaxLen, "domain")...)
	validators = append(validators, basedto.ValidateEmail(&req.Email, false, "email")...)
	validators = append(validators, basedto.ValidateStrIn(&req.KeyType, false,
		base.AllSSLKeyTypes, "keyType")...)
	return apperrors.NewValidationErrors(vld.Validate(validators...))
}

type ObtainDomainSSLResp struct {
	Meta *basedto.Meta         `json:"meta"`
	Data *basedto.ObjectIDResp `json:"data"`
}
