package sslcertdto

import (
	"strings"
	"time"

	vld "github.com/tiendc/go-validator"
	"github.com/tiendc/gofn"

	"github.com/hivepaas/hivepaas/hivepaas_app/apperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/basedto"
	"github.com/hivepaas/hivepaas/hivepaas_app/config"
	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/domainhelper"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/timeutil"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/settings"
)

const (
	keyMaxLen = 20000
)

type CreateSSLCertReq struct {
	settings.CreateSettingReq
	*SSLCertBaseReq
}

type SSLCertBaseReq struct {
	CertType     base.SSLCertType                  `json:"certType"`
	Provider     basedto.ObjectIDReq               `json:"provider"`
	Domain       string                            `json:"domain"`
	Certificate  string                            `json:"certificate"`
	PrivateKey   string                            `json:"privateKey"`
	KeyType      base.SSLKeyType                   `json:"keyType"`
	ValidPeriod  timeutil.Duration                 `json:"validPeriod"`
	Email        string                            `json:"email"`
	AutoRenew    bool                              `json:"autoRenew"`
	AcmeProvider basedto.ObjectIDReq               `json:"acmeProvider"`
	ExpireAt     time.Time                         `json:"expireAt"`
	NotifyFrom   time.Time                         `json:"notifyFrom"`
	Notification *basedto.BaseEventNotificationReq `json:"notification"`
}

func (req *SSLCertBaseReq) ToEntity() *entity.SSLCert {
	return &entity.SSLCert{
		CertType:     req.CertType,
		Provider:     entity.ObjectID{ID: req.Provider.ID},
		Domain:       req.Domain,
		Certificate:  req.Certificate,
		PrivateKey:   entity.NewEncryptedField(req.PrivateKey),
		KeyType:      req.KeyType,
		ValidPeriod:  req.ValidPeriod,
		Email:        req.Email,
		AutoRenew:    req.AutoRenew,
		AcmeProvider: entity.ObjectID{ID: req.AcmeProvider.ID},
		ExpireAt:     req.ExpireAt,
		NotifyFrom:   req.NotifyFrom,
		Notification: req.Notification.ToEntity(),
	}
}

func (req *SSLCertBaseReq) modifyRequest() error {
	req.Domain = strings.TrimSpace(req.Domain)
	req.Email = strings.TrimSpace(req.Email)
	req.KeyType = gofn.Coalesce(req.KeyType, base.SSLKeyTypeDefault)
	switch req.CertType {
	case base.SSLCertTypeLetsEncrypt:
		// Do nothing
	case base.SSLCertTypeZeroSSL:
		// Do nothing
	case base.SSLCertTypeGoogleTrust:
		// Do nothing
	case base.SSLCertTypeSelfSigned:
		req.ValidPeriod = gofn.Coalesce(req.ValidPeriod, timeutil.Duration(base.SSLSelfSignedValidPeriodDefault))
	case base.SSLCertTypeCustom:
		req.AutoRenew = false
	}
	return nil
}

func (req *SSLCertBaseReq) validate(field string) (res []vld.Validator) {
	if field != "" {
		field += "."
	}

	cfg := config.Current
	requireCert := req.CertType == base.SSLCertTypeCustom
	requireProvider := req.CertType == base.SSLCertTypeZeroSSL || req.CertType == base.SSLCertTypeGoogleTrust
	requireAcmeProvider := strings.HasPrefix(req.Domain, "*.") && req.CertType != base.SSLCertTypeCustom &&
		req.CertType != base.SSLCertTypeSelfSigned

	res = append(res, basedto.ValidateStrIn(&req.CertType, true, base.AllSSLCertTypes, field+"certType")...)
	res = append(res, basedto.ValidateObjectIDReq(&req.Provider, requireProvider, field+"provider")...)
	res = append(res, basedto.ValidateDomain(&req.Domain, true, base.DomainNameMaxLen, true, field+"domain")...)
	res = append(res, basedto.ValidateStr(&req.Certificate, requireCert, 1, keyMaxLen, field+"certificate")...)
	res = append(res, basedto.ValidateStr(&req.PrivateKey, requireCert, 1, keyMaxLen, field+"privateKey")...)
	res = append(res, basedto.ValidateEmail(&req.Email, false, field+"email")...)
	res = append(res, basedto.ValidateObjectIDReq(&req.AcmeProvider, requireAcmeProvider, field+"acmeProvider")...)

	res = append(res, vld.Must(domainhelper.IsSubdomainOrEqual(cfg.RootDomain, req.Domain)).OnError(
		vld.SetField(field+"domain", nil),
		vld.SetCustomKey("ERR_VLD_SUBDOMAIN_REQUIRED"),
		vld.SetParam("Domain", cfg.RootDomain),
	))

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
