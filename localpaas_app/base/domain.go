package base

type SSLCertType string

const (
	SSLCertTypeLetsEncrypt SSLCertType = "letsencrypt"
	SSLCertTypeCustom      SSLCertType = "custom"
	SSLCertTypeSelfSigned  SSLCertType = "self-signed"
)

var (
	AllSSLCertTypes = []SSLCertType{SSLCertTypeLetsEncrypt, SSLCertTypeCustom, SSLCertTypeSelfSigned}
)
