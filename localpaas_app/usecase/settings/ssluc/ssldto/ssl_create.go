package ssldto

import (
	"strings"
	"time"

	vld "github.com/tiendc/go-validator"

	"github.com/localpaas/localpaas/localpaas_app/apperrors"
	"github.com/localpaas/localpaas/localpaas_app/base"
	"github.com/localpaas/localpaas/localpaas_app/basedto"
	"github.com/localpaas/localpaas/localpaas_app/entity"
	"github.com/localpaas/localpaas/localpaas_app/pkg/timeutil"
	"github.com/localpaas/localpaas/localpaas_app/usecase/settings"
)

const (
	keyMaxLen      = 10000
	providerMaxLen = 100
)

type CreateSSLReq struct {
	settings.CreateSettingReq
	*SSLBaseReq
}

type SSLBaseReq struct {
	Name        string            `json:"name"`
	Domain      string            `json:"domain"`
	Certificate string            `json:"certificate"`
	PrivateKey  string            `json:"privateKey"`
	KeySize     int               `json:"keySize"`
	Provider    base.SSLProvider  `json:"provider"`
	Email       string            `json:"email"`
	AutoRenew   bool              `json:"autoRenew"`
	ExpireAt    time.Time         `json:"expireAt"`
	NotifyWhen  timeutil.Duration `json:"notifyWhen"`
}

func (req *SSLBaseReq) ToEntity() *entity.SSL {
	return &entity.SSL{
		Domain:      req.Domain,
		Certificate: req.Certificate,
		PrivateKey:  entity.NewEncryptedField(req.PrivateKey),
		KeySize:     req.KeySize,
		Provider:    req.Provider,
		Email:       req.Email,
		AutoRenew:   req.AutoRenew,
		ExpireAt:    req.ExpireAt,
		NotifyWhen:  req.NotifyWhen,
	}
}

func (req *SSLBaseReq) modifyRequest() error {
	req.Name = strings.TrimSpace(req.Name)
	return nil
}

func (req *SSLBaseReq) validate(field string) (res []vld.Validator) {
	if field != "" {
		field += "."
	}
	res = append(res, basedto.ValidateStr(&req.Name, true, 1, base.SettingNameMaxLen, field+"name")...)
	res = append(res, basedto.ValidateStr(&req.Certificate, true, 1, keyMaxLen, field+"certificate")...)
	res = append(res, basedto.ValidateStr(&req.PrivateKey, true, 1, keyMaxLen, field+"privateKey")...)
	res = append(res, basedto.ValidateStr(&req.Provider, false, 1, providerMaxLen, field+"provider")...)
	res = append(res, basedto.ValidateEmail(&req.Email, false, field+"email")...)
	return res
}

func NewCreateSSLReq() *CreateSSLReq {
	return &CreateSSLReq{}
}

func (req *CreateSSLReq) ModifyRequest() error {
	return req.modifyRequest()
}

// Validate implements interface basedto.ReqValidator
func (req *CreateSSLReq) Validate() apperrors.ValidationErrors {
	var validators []vld.Validator
	validators = append(validators, req.validate("")...)
	return apperrors.NewValidationErrors(vld.Validate(validators...))
}

type CreateSSLResp struct {
	Meta *basedto.Meta         `json:"meta"`
	Data *basedto.ObjectIDResp `json:"data"`
}
