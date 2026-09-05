package hpappservice

import (
	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/secrethelper"
)

// appSecretMinLen is longer than the default: the app secret is typed once by an
// operator and protects every stored secret, so it is worth more than a login.
const appSecretMinLen = 12

var (
	SecretRequirements = secrethelper.SecretStrengthRequirements{
		MinLen:             appSecretMinLen,
		MaxLen:             secrethelper.DefaultSecretMaxLen,
		RequiredLowercases: secrethelper.DefaultSecretRequiredLowercases,
		RequiredUppercases: secrethelper.DefaultSecretRequiredUppercases,
		RequiredDigits:     secrethelper.DefaultSecretRequiredDigits,
		RequiredSpecials:   secrethelper.DefaultSecretRequiredSpecials,
		MaxSimilarRun:      secrethelper.DefaultSecretMaxSimilarRun,
		MaxSequenceRun:     secrethelper.DefaultSecretMaxSequenceRun,
	}
)

type AppReleaseInfo struct {
	Stable *ReleaseInfo `json:"stable"`
	Beta   *ReleaseInfo `json:"beta"`
}

type ReleaseInfo struct {
	base.ReleaseInfo

	// System specific flag
	CanUpdate bool `json:"canUpdate"`
}
