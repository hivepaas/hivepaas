package base

import "time"

type SSLCertType string

const (
	SSLCertTypeLetsEncrypt SSLCertType = "letsencrypt"
	SSLCertTypeCustom      SSLCertType = "custom"
	SSLCertTypeSelfSigned  SSLCertType = "self-signed"
)

var (
	AllSSLCertTypes = []SSLCertType{SSLCertTypeLetsEncrypt, SSLCertTypeCustom, SSLCertTypeSelfSigned}
)

type SSLKeyType string

const (
	SSLKeyTypeECP256  SSLKeyType = "ec-p256"
	SSLKeyTypeECP384  SSLKeyType = "ec-p384"
	SSLKeyTypeECP521  SSLKeyType = "ec-p521"
	SSLKeyTypeRSA2048 SSLKeyType = "rsa-2048"
	SSLKeyTypeRSA3072 SSLKeyType = "rsa-3072"
	SSLKeyTypeRSA4096 SSLKeyType = "rsa-4096"
)

var (
	AllSSLKeyTypes = []SSLKeyType{SSLKeyTypeECP256, SSLKeyTypeECP384, SSLKeyTypeECP521,
		SSLKeyTypeRSA2048, SSLKeyTypeRSA3072, SSLKeyTypeRSA4096}
)

const (
	SSLKeyTypeDefault = SSLKeyTypeECP256

	SSLSelfSignedValidPeriodDefault   = time.Hour * 24 * 365 // 365 days
	SSLSelfSignedRenewalPeriodDefault = time.Hour * 24 * 30  // 30 days

	DomainNameMaxLen = 100

	LetsEncryptExpirationFromFirstRenewableDate = time.Hour * 24 * 30 // 30 days
)
