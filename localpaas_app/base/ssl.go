package base

import (
	"github.com/localpaas/localpaas/localpaas_app/pkg/timeutil"
)

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

	DomainNameMaxLen = 100

	SSLExpirationFromFirstRenewableDate = timeutil.Day * 30

	ZeroSSLACMEURL  = "https://acme.zerossl.com/v2/DV90"
	GoogleTSACMEURL = "https://dv.acme-v02.api.pki.goog/directory"
)
