// Package license verifies a license key and works out what it puts in force.
//
// A license key is a releasesig envelope: the payload, base64, signed with an
// ed25519 key and an ML-DSA-65 key, both required. It is signed with key pairs
// of its own and under a context of its own, so a release key never makes a
// license key and a license key never makes a release.json. The portal signs
// keys unattended; tools/licensesign signs them by hand, for support and for
// development.
//
// The package is pure: it reads no database, no docker and no config. What it
// needs from them - the trusted keys, whether development keys count, this
// installation's ID, the time - is passed in, so the service decides it and tests
// can set it.
//
// A payload carries the customer's email. Nothing here puts a payload field in
// an error, and nothing that holds a payload (Payload, Verified, State) should be
// logged.
package license

import (
	"slices"
	"time"

	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/releasesig"
)

const (
	// Context is the signing context of both algorithms for a license key.
	Context = "hivepaas-license-v1"

	// PayloadVersion is the payload format this binary reads. A key of another
	// version is refused rather than half-read.
	PayloadVersion = 1

	// maxGraceDays bounds a key's grace, so a mistyped key cannot hold for years
	// past its period.
	maxGraceDays = 365

	day = 24 * time.Hour
)

// Purpose opens a license key: under Context, refused as ErrLicenseInvalid, with
// a bad compiled-in key set reported as ErrLicenseSigningKeyInvalid.
var Purpose = releasesig.Purpose{
	Context:      Context,
	ErrSignature: hperrors.ErrLicenseInvalid,
	ErrKeys:      hperrors.ErrLicenseSigningKeyInvalid,
}

// Payload is what a license key says. tools/licensesign writes the same fields.
//
// Limits is by limit name, as written: a name this binary does not know is
// passed over, so a key from a newer portal still works here. A limit left out
// is no limit.
type Payload struct {
	Version        int              `json:"version"`
	LicenseID      string           `json:"licenseId"`
	SubscriptionID string           `json:"subscriptionId,omitempty"`
	Customer       Customer         `json:"customer"`
	Plan           base.LicensePlan `json:"plan"`
	Limits         map[string]int   `json:"limits,omitempty"`
	Installations  []string         `json:"installations"`
	IssuedAt       time.Time        `json:"issuedAt"`
	ExpiresAt      time.Time        `json:"expiresAt"`
	GraceDays      int              `json:"graceDays"`
	Trial          bool             `json:"trial,omitempty"`
}

// Customer is who a key was issued to. Email is never logged, put in an error,
// or sent back by the API.
type Customer struct {
	ID    string `json:"id,omitempty"`
	Name  string `json:"name,omitempty"`
	Email string `json:"email,omitempty"`
}

// GraceEndsAt is when the key stops holding: its graceDays after its expiresAt.
func (p *Payload) GraceEndsAt() time.Time {
	return p.ExpiresAt.Add(time.Duration(p.GraceDays) * day)
}

// ForInstallation reports whether the key names the installation.
func (p *Payload) ForInstallation(installationID string) bool {
	return slices.Contains(p.Installations, installationID)
}

// validate checks what a signature cannot: that the payload is one this binary
// can apply. It runs only on a payload whose signatures verified.
//
// The details name what is wrong and never a value the customer gave.
func (p *Payload) validate() error {
	if p.Version != PayloadVersion {
		return invalid("payload version %d is not one this HivePaaS reads", p.Version)
	}
	if p.LicenseID == "" {
		return invalid("no license id")
	}
	switch p.Plan {
	case base.LicensePlanComb, base.LicensePlanHive, base.LicensePlanColony, base.LicensePlanApiary:
	case base.LicensePlanBee:
		// Bee is what an installation runs without a key.
		return invalid("plan %q is never in a key", p.Plan)
	default:
		return invalid("plan %q is not one this HivePaaS knows", p.Plan)
	}
	if len(p.Installations) == 0 {
		return invalid("no installation")
	}
	if slices.Contains(p.Installations, "") {
		return invalid("an empty installation ID")
	}
	if p.IssuedAt.IsZero() || p.ExpiresAt.IsZero() {
		return invalid("issuedAt and expiresAt are both required")
	}
	if !p.ExpiresAt.After(p.IssuedAt) {
		return invalid("expiresAt is not after issuedAt")
	}
	if p.GraceDays < 0 || p.GraceDays > maxGraceDays {
		return invalid("graceDays is not between 0 and %d", maxGraceDays)
	}
	for name, value := range p.Limits {
		if value < 0 {
			return invalid("limit %q is negative", name)
		}
	}
	return nil
}

func invalid(format string, args ...any) error {
	return hperrors.Wrap(hperrors.ErrLicenseInvalid).WithExtraDetail(format, args...)
}
