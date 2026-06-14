package base

import (
	"github.com/localpaas/localpaas/localpaas_app/pkg/timeutil"
)

type SSLProvider string

const (
	SSLProviderGoogleTrust SSLProvider = "googletrust"
	SSLProviderLetsEncrypt SSLProvider = "letsencrypt"
	SSLProviderZeroSSL     SSLProvider = "zerossl"
)

var (
	AllSSLProviders = []SSLProvider{SSLProviderGoogleTrust, SSLProviderLetsEncrypt, SSLProviderZeroSSL}
)

type SSLCertType string

const (
	SSLCertTypeGoogleTrust SSLCertType = SSLCertType(SSLProviderGoogleTrust)
	SSLCertTypeLetsEncrypt SSLCertType = SSLCertType(SSLProviderLetsEncrypt)
	SSLCertTypeZeroSSL     SSLCertType = SSLCertType(SSLProviderZeroSSL)
	SSLCertTypeCustom      SSLCertType = "custom"
	SSLCertTypeSelfSigned  SSLCertType = "self-signed"
)

var (
	AllSSLCertTypes = []SSLCertType{SSLCertTypeGoogleTrust, SSLCertTypeLetsEncrypt, SSLCertTypeZeroSSL,
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
	SSLKeyTypeRSA8192 = SSLKeyType(PrivateKeyTypeRSA8192)
)

var (
	AllSSLKeyTypes = []SSLKeyType{SSLKeyTypeECP256, SSLKeyTypeECP384, SSLKeyTypeECP521,
		SSLKeyTypeRSA2048, SSLKeyTypeRSA3072, SSLKeyTypeRSA4096, SSLKeyTypeRSA8192}
)

const (
	SSLKeyTypeDefault = SSLKeyTypeECP256

	SSLSelfSignedValidPeriodDefault   = timeutil.Day * 365
	SSLSelfSignedRenewalPeriodDefault = timeutil.Day * 30

	SSLExpirationFromFirstRenewableDate = timeutil.Day * 30
)
