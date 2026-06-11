package base

import "time"

type SSLProvider string

const (
	SSLProviderLetsEncrypt SSLProvider = "letsencrypt"
	SSLProviderZeroSSL     SSLProvider = "zerossl"
	SSLProviderGoogleTS    SSLProvider = "googlets"
)

var (
	AllSSLProviders = []SSLProvider{SSLProviderLetsEncrypt, SSLProviderZeroSSL, SSLProviderGoogleTS}
)

type SSLCertType string

const (
	SSLCertTypeLetsEncrypt SSLCertType = SSLCertType(SSLProviderLetsEncrypt)
	SSLCertTypeZeroSSL     SSLCertType = SSLCertType(SSLProviderZeroSSL)
	SSLCertTypeGoogleTS    SSLCertType = SSLCertType(SSLProviderGoogleTS)
	SSLCertTypeCustom      SSLCertType = "custom"
	SSLCertTypeSelfSigned  SSLCertType = "self-signed"
)

var (
	AllSSLCertTypes = []SSLCertType{SSLCertTypeLetsEncrypt, SSLCertTypeZeroSSL, SSLCertTypeGoogleTS,
		SSLCertTypeCustom, SSLCertTypeSelfSigned}
)

type SSLKeyType string

const (
	SSLKeyTypeECP256  = SSLKeyType(PrivateKeyTypeECP256)
	SSLKeyTypeECP384  = SSLKeyType(PrivateKeyTypeECP384)
	SSLKeyTypeECP521  = SSLKeyType(PrivateKeyTypeECP521)
	SSLKeyTypeRSA2048 = SSLKeyType(PrivateKeyTypeRSA2048)
	SSLKeyTypeRSA3072 = SSLKeyType(PrivateKeyTypeRSA3072)
	SSLKeyTypeRSA4096 = SSLKeyType(PrivateKeyTypeRSA4096)
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
