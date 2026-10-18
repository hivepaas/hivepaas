package license

import (
	"maps"
	"time"

	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/translation"
)

// EvalInput is what Evaluate decides from.
type EvalInput struct {
	// KeyText is the stored key, empty when there is none.
	KeyText string
	// Verified and VerifyErr are Verify's answer for KeyText. The caller keeps
	// them between calls, as the signatures of a key that has not changed need
	// not be verified again.
	Verified  *Verified
	VerifyErr error
	// InstallationID is this installation's: its swarm cluster ID.
	InstallationID string
	// Now is the time to evaluate at, as EffectiveNow gives it.
	Now time.Time
	// MovedAt is when the portal said the key was moved to another installation,
	// zero when it has not.
	MovedAt time.Time
}

// State is the license in force.
//
// Key holds the key's payload, and with it the customer's email: never log a
// State.
type State struct {
	Status base.LicenseStatus
	// Plan is the plan in force: the key's, or Bee when the key puts nothing in
	// force.
	Plan base.LicensePlan
	// Limits are the limits in force. They are the State's own, never shared.
	Limits Limits
	// Features are the features in force: the key's that this binary knows,
	// while the key holds. Bee grants none; a development installation may be
	// given some whatever its key (State.WithFeatures).
	Features []base.LicenseFeature
	// Key is the key's payload whenever its signatures verified, also when it is
	// not in force, as for another installation or past its grace.
	Key *Payload
	// Development is a key signed by the development keys.
	Development bool
	// Reason says why the key is invalid, in the words the API would refuse it
	// with.
	Reason         string
	InstallationID string
	// GraceEndsAt is when a key for this installation stops holding: its grace
	// after it expires, or after it was moved if that comes first.
	GraceEndsAt time.Time
	MovedAt     time.Time
}

// Evaluate works out the license in force from the stored key. A key puts its
// plan and limits in force only while it is for this installation and within its
// grace; otherwise the installation is on Bee, which only stops creating.
//
// A key's issuedAt does not hold it back: an installation whose clock is a few
// minutes behind the portal's would otherwise refuse the key it was just sold.
func Evaluate(in *EvalInput) *State {
	state := &State{
		Status:         base.LicenseStatusNone,
		Plan:           base.LicensePlanBee,
		Limits:         maps.Clone(BeeLimits),
		InstallationID: in.InstallationID,
		MovedAt:        in.MovedAt,
	}
	if in.KeyText == "" {
		return state
	}
	if in.VerifyErr != nil || in.Verified == nil {
		state.Status = base.LicenseStatusInvalid
		state.Reason = reasonOf(in.VerifyErr)
		return state
	}

	key := in.Verified.Payload
	state.Key = key
	state.Development = in.Verified.Development
	if !key.ForInstallation(in.InstallationID) {
		state.Status = base.LicenseStatusOtherInstallation
		return state
	}

	state.GraceEndsAt = key.GraceEndsAt()
	switch {
	case !in.MovedAt.IsZero():
		// An installation left on a moved key keeps it for the key's grace,
		// counted from the move, as long as a lapsed key would be kept - or less,
		// when the key's own grace ends sooner.
		state.Status = base.LicenseStatusMoved
		movedGraceEndsAt := in.MovedAt.Add(time.Duration(key.GraceDays) * day)
		if movedGraceEndsAt.Before(state.GraceEndsAt) {
			state.GraceEndsAt = movedGraceEndsAt
		}
	case in.Now.Before(key.ExpiresAt):
		state.Status = base.LicenseStatusValid
		if key.Trial {
			state.Status = base.LicenseStatusTrial
		}
	case in.Now.Before(state.GraceEndsAt):
		state.Status = base.LicenseStatusGrace
	default:
		state.Status = base.LicenseStatusExpired
	}
	if in.Now.Before(state.GraceEndsAt) {
		state.Plan = key.Plan
		state.Limits = LimitsOf(key)
		state.Features = FeaturesOf(key)
	}
	return state
}

// reasonOf is the detail the API would refuse a key with for err.
func reasonOf(err error) string {
	if err == nil {
		return ""
	}
	info, _ := hperrors.ParseError(err, translation.GetDefaultLang())
	return info.Detail
}
