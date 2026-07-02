package base

type MFAType string

const (
	MFATypeTOTP  = MFAType("totp")
	MFATypeEmail = MFAType("email")
)
