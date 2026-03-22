package base

import "time"

const (
	SSLKeySizeDefault = 2048

	DomainNameMaxLen = 100

	LetsEncryptExpirationFromFirstRenewableDate = time.Hour * 24 * 30 // 30 days
)

var (
	SSLKeySizesAllowed = []int{2048, 3072, 4096}
)
