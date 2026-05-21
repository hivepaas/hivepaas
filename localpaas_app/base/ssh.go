package base

type PrivateKeyType string

const (
	PrivateKeyTypeEd25519 PrivateKeyType = "ed25519"
	PrivateKeyTypeECP256  PrivateKeyType = "ec-p256"
	PrivateKeyTypeECP384  PrivateKeyType = "ec-p384"
	PrivateKeyTypeECP521  PrivateKeyType = "ec-p521"
	PrivateKeyTypeRSA2048 PrivateKeyType = "rsa-2048"
	PrivateKeyTypeRSA3072 PrivateKeyType = "rsa-3072"
	PrivateKeyTypeRSA4096 PrivateKeyType = "rsa-4096"
	PrivateKeyTypeRSA8192 PrivateKeyType = "rsa-8192"
)

var (
	AllPrivateKeyTypes = []PrivateKeyType{PrivateKeyTypeEd25519,
		PrivateKeyTypeECP256, PrivateKeyTypeECP384, PrivateKeyTypeECP521,
		PrivateKeyTypeRSA2048, PrivateKeyTypeRSA3072, PrivateKeyTypeRSA4096, PrivateKeyTypeRSA8192}
)
