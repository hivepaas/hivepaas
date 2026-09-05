package secrethelper

import (
	"unicode"

	"github.com/tiendc/gofn"

	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
)

const (
	DefaultSecretMinLen             = 8
	DefaultSecretMaxLen             = 50
	DefaultSecretRequiredLowercases = 1
	DefaultSecretRequiredUppercases = 1
	DefaultSecretRequiredDigits     = 1
	DefaultSecretRequiredSpecials   = 1

	// DefaultSecretMaxSimilarRun is the longest stretch of characters a new secret may
	// share with one it replaces. Past that the "new" secret is usually the old one
	// with a digit bumped, which is what this rule exists to catch.
	DefaultSecretMaxSimilarRun = 3

	// DefaultSecretMaxSequenceRun is the longest run of consecutive ("1234", "dcba"),
	// repeated ("aaaa") or keyboard-adjacent ("qwer") characters a secret may contain.
	// Three leaves the "abc" that happens to land inside an otherwise strong secret
	// alone while rejecting the runs people actually build passwords out of. Set it to
	// 2 to reject "123" and "321" as well, at the cost of more honest refusals.
	DefaultSecretMaxSequenceRun = 3
)

var (
	specialCharset  = []rune("!@#$%^&*()_+-=[]{}|;':\",./<>?`~\\")
	mapSpecialChars = gofn.MapSliceToMap(specialCharset, func(k rune) (rune, struct{}) {
		return k, struct{}{}
	})
)

// SecretStrengthRequirements is the policy one secret is checked against.
//
// The six count fields keep the sentinel the old positional API used: a negative
// value means "use the default", zero means "do not require any". MaxSimilarRun and
// MaxSequenceRun have no meaningful zero - a maximum run of zero would reject every
// secret - so for those anything at or below zero means "use the default".
type SecretStrengthRequirements struct {
	MinLen             int
	MaxLen             int
	RequiredLowercases int
	RequiredUppercases int
	RequiredDigits     int
	RequiredSpecials   int

	// MaxSimilarRun caps how much of a previous secret may survive into the new one.
	MaxSimilarRun int
	// MaxSequenceRun caps runs of sequential, repeated or keyboard-adjacent characters.
	MaxSequenceRun int

	// PrevSecrets are the plaintext secrets the new one is compared against, each read
	// forwards and backwards.
	//
	// Only secrets the caller already holds in plaintext belong here. For a user
	// password that is exactly one - the current password sent in the same request -
	// because the stored form is a hash, and a hash tells you nothing about how much
	// of it a new password resembles. A caller that already keeps its secrets
	// recoverable, as the backup repository does, can pass more than one; that is a
	// decision about what to store, made where the storing happens, not here.
	PrevSecrets []string
}

// DefaultRequirements is the policy a caller gets by passing nil: every rule at its
// default. It saves spelling out six -1s at the call site.
func DefaultRequirements() SecretStrengthRequirements {
	return SecretStrengthRequirements{
		MinLen:             -1,
		MaxLen:             -1,
		RequiredLowercases: -1,
		RequiredUppercases: -1,
		RequiredDigits:     -1,
		RequiredSpecials:   -1,
	}
}

func (req *SecretStrengthRequirements) GetRequirementMap() map[string]any {
	return map[string]any{
		"MinLen":   req.MinLen,
		"MaxLen":   req.MaxLen,
		"Lowers":   req.RequiredLowercases,
		"Uppers":   req.RequiredUppercases,
		"Digits":   req.RequiredDigits,
		"Specials": req.RequiredSpecials,
	}
}

// normalized returns a copy with every unset field replaced by its default, so the
// checks read the values once instead of repeating the fallbacks. The value receiver
// is what makes the copy; the caller's policy is never modified.
func (req SecretStrengthRequirements) normalized() SecretStrengthRequirements {
	if req.MinLen < 0 {
		req.MinLen = DefaultSecretMinLen
	}
	if req.MaxLen < 0 {
		req.MaxLen = DefaultSecretMaxLen
	}
	if req.RequiredLowercases < 0 {
		req.RequiredLowercases = DefaultSecretRequiredLowercases
	}
	if req.RequiredUppercases < 0 {
		req.RequiredUppercases = DefaultSecretRequiredUppercases
	}
	if req.RequiredDigits < 0 {
		req.RequiredDigits = DefaultSecretRequiredDigits
	}
	if req.RequiredSpecials < 0 {
		req.RequiredSpecials = DefaultSecretRequiredSpecials
	}
	if req.MaxSimilarRun <= 0 {
		req.MaxSimilarRun = DefaultSecretMaxSimilarRun
	}
	if req.MaxSequenceRun <= 0 {
		req.MaxSequenceRun = DefaultSecretMaxSequenceRun
	}
	return req
}

// ValidateStrength checks a secret against the strength requirements. A nil policy
// means the defaults.
//
// No error carries the secret, a previous secret, or the run that matched: the params
// say what the limit was, never what tripped it.
func ValidateStrength(secret string, requirements *SecretStrengthRequirements) error {
	policy := DefaultRequirements()
	if requirements != nil {
		policy = *requirements
	}
	policy = policy.normalized()

	chars := []rune(secret)
	if len(chars) < policy.MinLen || len(chars) > policy.MaxLen {
		return hperrors.Wrap(hperrors.ErrPasswordNotMeetRequirements).
			WithParams(policy.GetRequirementMap()).
			WithMsgLog("incorrect length: %d", len(chars))
	}

	lowers, uppers, digits, specials := countCharClasses(chars)
	if lowers < policy.RequiredLowercases || uppers < policy.RequiredUppercases ||
		digits < policy.RequiredDigits || specials < policy.RequiredSpecials {
		return hperrors.Wrap(hperrors.ErrPasswordNotMeetRequirements).
			WithParams(policy.GetRequirementMap()).
			WithMsgLog("lowers %d, uppers %d, digits %d, specials %d", lowers, uppers, digits, specials)
	}

	if run := longestWeakRun(chars); run > policy.MaxSequenceRun {
		return hperrors.Wrap(hperrors.ErrPasswordHasWeakSequence).
			WithParams(map[string]any{"MaxRun": policy.MaxSequenceRun}).
			WithMsgLog("weak run of %d characters", run)
	}

	for _, prev := range policy.PrevSecrets {
		if prev == "" {
			continue
		}
		if run := longestSharedRun(chars, []rune(prev)); run > policy.MaxSimilarRun {
			return hperrors.Wrap(hperrors.ErrPasswordTooSimilarToPrevious).
				WithParams(map[string]any{"MaxRun": policy.MaxSimilarRun}).
				WithMsgLog("shares a run of %d characters with a previous secret", run)
		}
	}

	return nil
}

// countCharClasses counts the four classes the requirements are expressed in.
//
// A rune that belongs to none of them - a space, a symbol outside specialCharset -
// counts towards nothing. Letting such runes fall through to "uppercase" would mean a
// single space satisfied the uppercase requirement.
func countCharClasses(chars []rune) (lowers, uppers, digits, specials int) {
	for _, r := range chars {
		switch {
		case unicode.IsDigit(r):
			digits++
		case gofn.MapContainKeys(mapSpecialChars, r):
			specials++
		case unicode.IsLower(r):
			lowers++
		case unicode.IsUpper(r):
			uppers++
		}
	}
	return lowers, uppers, digits, specials
}
