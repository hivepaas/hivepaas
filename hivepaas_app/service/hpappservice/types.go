package hpappservice

import (
	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/secrethelper"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/timeutil"
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
	// Current is the release this installation runs. It is not read from the
	// release file, which only knows what is published.
	Current *CurrentRelease `json:"current"`
	Stable  *ReleaseInfo    `json:"stable"`
	Beta    *ReleaseInfo    `json:"beta"`
}

// CurrentRelease is the release an installation runs, and the channel it
// follows.
type CurrentRelease struct {
	AppVersion  string        `json:"appVersion"`
	Channel     string        `json:"channel"`
	ReleaseDate timeutil.Date `json:"releaseDate"`
}

const (
	ReleaseChannelStable = "stable"
	ReleaseChannelBeta   = "beta"
)

// ReleaseRelation is where a published release stands against the one running.
type ReleaseRelation string

const (
	ReleaseNewer ReleaseRelation = "newer"
	ReleaseSame  ReleaseRelation = "same"
	// ReleaseOlder is a release behind the one running - a stable release while
	// running a later beta. Moving to it would be a downgrade, which the updater
	// does not do: migrations only run forward.
	ReleaseOlder ReleaseRelation = "older"
)

type ReleaseInfo struct {
	base.ReleaseInfo

	// CanUpdate is whether this installation may move to the release, which is
	// whether it is newer than the one running.
	CanUpdate bool            `json:"canUpdate"`
	Relation  ReleaseRelation `json:"relation"`
}
