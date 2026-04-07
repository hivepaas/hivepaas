package sslcertdto

import (
	"strings"
	"time"

	vld "github.com/tiendc/go-validator"

	"github.com/localpaas/localpaas/localpaas_app/apperrors"
	"github.com/localpaas/localpaas/localpaas_app/base"
	"github.com/localpaas/localpaas/localpaas_app/basedto"
	"github.com/localpaas/localpaas/localpaas_app/entity"
	"github.com/localpaas/localpaas/localpaas_app/usecase/settings"
)

const (
	keyMaxLen = 50000
)

type CreateSSLCertReq struct {
	settings.CreateSettingReq
	*SSLCertBaseReq
}

type SSLCertBaseReq struct {
	Name        string           `json:"name"`
	CertType    base.SSLCertType `json:"certType"`
	Domain      string           `json:"domain"`
	Certificate string           `json:"certificate"`
	PrivateKey  string           `json:"privateKey"`
	KeyType     base.SSLKeyType  `json:"keyType"`
	Email       string           `json:"email"`
	AutoRenew   bool             `json:"autoRenew"`
	ExpireAt    time.Time        `json:"expireAt"`
	NotifyFrom  time.Time        `json:"notifyFrom"`
}

func (req *SSLCertBaseReq) ToEntity() *entity.SSLCert {
	return &entity.SSLCert{
		CertType:    req.CertType,
		Domain:      req.Domain,
		Certificate: req.Certificate,
		PrivateKey:  entity.NewEncryptedField(req.PrivateKey),
		KeyType:     req.KeyType,
		Email:       req.Email,
		AutoRenew:   req.AutoRenew,
		ExpireAt:    req.ExpireAt,
		NotifyFrom:  req.NotifyFrom,
	}
}

func (req *SSLCertBaseReq) modifyRequest() error {
	req.Name = strings.TrimSpace(req.Name)
	return nil
}

func (req *SSLCertBaseReq) validate(field string) (res []vld.Validator) {
	if field != "" {
		field += "."
	}
	res = append(res, basedto.ValidateStr(&req.Name, true, 1, base.SettingNameMaxLen, field+"name")...)
	res = append(res, basedto.ValidateStrIn(&req.CertType, true, base.AllSSLCertTypes, field+"certType")...)
	res = append(res, basedto.ValidateStr(&req.Certificate, true, 1, keyMaxLen, field+"certificate")...)
	res = append(res, basedto.ValidateStr(&req.PrivateKey, true, 1, keyMaxLen, field+"privateKey")...)
	res = append(res, basedto.ValidateEmail(&req.Email, false, field+"email")...)
	return res
}

func NewCreateSSLCertReq() *CreateSSLCertReq {
	return &CreateSSLCertReq{}
}

func (req *CreateSSLCertReq) ModifyRequest() error {
	return req.modifyRequest()
}

// Validate implements interface basedto.ReqValidator
func (req *CreateSSLCertReq) Validate() apperrors.ValidationErrors {
	var validators []vld.Validator
	validators = append(validators, req.validate("")...)
	return apperrors.NewValidationErrors(vld.Validate(validators...))
}

type CreateSSLCertResp struct {
	Meta *basedto.Meta         `json:"meta"`
	Data *basedto.ObjectIDResp `json:"data"`
}
