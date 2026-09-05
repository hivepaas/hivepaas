package secrethelper

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
)

// lenientReqs turns off every rule except the one under test, so a case that is meant
// to fail on a sequence cannot pass or fail for want of an uppercase letter instead.
func lenientReqs() SecretStrengthRequirements {
	return SecretStrengthRequirements{
		MinLen:             1,
		MaxLen:             DefaultSecretMaxLen,
		RequiredLowercases: 0,
		RequiredUppercases: 0,
		RequiredDigits:     0,
		RequiredSpecials:   0,
		MaxSequenceRun:     DefaultSecretMaxLen, // effectively off
		MaxSimilarRun:      DefaultSecretMaxLen, // effectively off
	}
}

func Test_ValidateStrength_LengthAndCharClasses(t *testing.T) {
	errFail := hperrors.ErrPasswordNotMeetRequirements
	tests := []struct {
		name    string
		secret  string
		wantErr error
	}{
		{"too short", "Ab1!12", errFail},
		{"too long", "A" + strings.Repeat("q", DefaultSecretMaxLen-2) + "1!", errFail},
		{"no lowercase", "A1!A1!B2@", errFail},
		{"no uppercase", "a1!a1!b2@", errFail},
		{"no digit", "Aa!Aa!Bb@", errFail},
		{"no symbol", "Aa1Aa1Bb2", errFail},
		{"valid secret", "Aa1!Qm7#", nil},
	}

	reqs := DefaultRequirements()
	reqs.MaxSequenceRun = DefaultSecretMaxLen // isolate: only length and classes here
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateStrength(tt.secret, &reqs)
			if tt.wantErr != nil {
				assert.ErrorIs(t, err, tt.wantErr)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

// A nil policy has to mean the defaults, not a zero policy that requires nothing.
func Test_ValidateStrength_NilRequirementsUsesDefaults(t *testing.T) {
	assert.ErrorIs(t, ValidateStrength("short", nil), hperrors.ErrPasswordNotMeetRequirements)
	assert.NoError(t, ValidateStrength("Aa1!Qm7#", nil))
}

// The uppercase requirement has to be read from its own field. It was briefly taken
// from RequiredLowercases, which no default-valued policy could reveal because both
// defaults are 1.
func Test_ValidateStrength_UppercaseUsesItsOwnField(t *testing.T) {
	reqs := lenientReqs()
	reqs.RequiredUppercases = 2

	assert.ErrorIs(t, ValidateStrength("Aqm7xzk", &reqs), hperrors.ErrPasswordNotMeetRequirements)
	assert.NoError(t, ValidateStrength("AQm7xzk", &reqs))

	// And the lowercase requirement must not be read from the uppercase field.
	reqs = lenientReqs()
	reqs.RequiredLowercases = 2
	assert.ErrorIs(t, ValidateStrength("AQMXZ7K", &reqs), hperrors.ErrPasswordNotMeetRequirements)
	assert.NoError(t, ValidateStrength("AQMxz7K", &reqs))
}

// A rune in no class - a space, a symbol outside the special set - must not count
// towards the uppercase requirement.
func Test_ValidateStrength_UnclassifiedRunesCountForNothing(t *testing.T) {
	reqs := lenientReqs()
	reqs.RequiredUppercases = 1
	assert.ErrorIs(t, ValidateStrength("abc def", &reqs), hperrors.ErrPasswordNotMeetRequirements)
}

func Test_ValidateStrength_WeakSequences(t *testing.T) {
	tests := []struct {
		name   string
		secret string
		reject bool
	}{
		{"ascending digits", "xk1234mq", true},
		{"descending digits", "xk654321mq", true},
		{"ascending letters", "xkabcdmq", true},
		{"descending letters", "xkdcbamq", true},
		{"repeated character", "xkaaaamq", true},
		{"repeated character mixed case", "xkAaAamq", true},
		{"keyboard row forward", "xzqwermq", true},
		{"keyboard row backward", "xzrewqmq", true},
		{"keyboard home row", "pmasdfpm", true},
		{"three is allowed", "xk123mq9", false},
		{"three letters allowed", "xkabcmq9", false},
		{"alternating is not a run", "xk1213mq", false},
		{"aba is not a run", "xkababmq", false},
		{"unrelated characters", "xk7q2m9z", false},
	}

	reqs := lenientReqs()
	reqs.MaxSequenceRun = DefaultSecretMaxSequenceRun
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateStrength(tt.secret, &reqs)
			if tt.reject {
				assert.ErrorIs(t, err, hperrors.ErrPasswordHasWeakSequence)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

// The threshold has to be honored, so a deployment can reject "123" by lowering it.
func Test_ValidateStrength_SequenceThresholdIsConfigurable(t *testing.T) {
	reqs := lenientReqs()
	reqs.MaxSequenceRun = 2
	assert.ErrorIs(t, ValidateStrength("xk123mq9", &reqs), hperrors.ErrPasswordHasWeakSequence)
	assert.NoError(t, ValidateStrength("xk12mq9p", &reqs))
}

func Test_ValidateStrength_SimilarityToPreviousSecret(t *testing.T) {
	tests := []struct {
		name   string
		secret string
		prev   string
		reject bool
	}{
		{"digit bumped at the end", "Myp@ssw0rd2", "Myp@ssw0rd1", true},
		{"digit bumped at the front", "2Myp@ssw0rd", "1Myp@ssw0rd", true},
		{"case changed only", "MYP@SSW0RD", "myp@ssw0rd", true},
		{"four shared characters", "zqvxpass9", "pass1word", true},
		{"three shared characters is allowed", "zqvxpas91", "pas1word7", false},
		{"unrelated secret", "zqvxm793", "Myp@ssw0rd1", false},
		{"reversed whole secret", "1dr0wss@pyM", "Myp@ssw0rd1", true},
		{"reversed fragment", "zqvx0wss@p9", "Myp@ssw0rd1", true},
		{"reversed and recased", "1DR0WSS@PYM", "myp@ssw0rd1", true},
		{"reversed three characters is allowed", "zqvxdr09m7", "Myp@ssw0rd1", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			reqs := lenientReqs()
			reqs.MaxSimilarRun = DefaultSecretMaxSimilarRun
			reqs.PrevSecrets = []string{tt.prev}

			err := ValidateStrength(tt.secret, &reqs)
			if tt.reject {
				assert.ErrorIs(t, err, hperrors.ErrPasswordTooSimilarToPrevious)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func Test_ValidateStrength_SimilarityEdgeCases(t *testing.T) {
	reqs := lenientReqs()
	reqs.MaxSimilarRun = DefaultSecretMaxSimilarRun

	// No previous secret to compare against: the rule simply does not apply.
	reqs.PrevSecrets = nil
	assert.NoError(t, ValidateStrength("Myp@ssw0rd1", &reqs))

	// An empty previous secret must be skipped, not treated as matching everything.
	reqs.PrevSecrets = []string{""}
	assert.NoError(t, ValidateStrength("Myp@ssw0rd1", &reqs))

	// Every entry is checked, not just the first.
	reqs.PrevSecrets = []string{"zzzzzzzz", "Myp@ssw0rd1"}
	assert.ErrorIs(t, ValidateStrength("Myp@ssw0rd2", &reqs),
		hperrors.ErrPasswordTooSimilarToPrevious)
}

// The policy the caller owns must not be modified by validating against it, or a
// package-level policy would pick up the previous secrets of whoever ran last.
func Test_ValidateStrength_DoesNotMutateCallerPolicy(t *testing.T) {
	reqs := SecretStrengthRequirements{MinLen: -1, MaxLen: -1}
	_ = ValidateStrength("Aa1!Qm7#", &reqs)

	assert.Equal(t, -1, reqs.MinLen)
	assert.Equal(t, -1, reqs.MaxLen)
	assert.Equal(t, 0, reqs.MaxSequenceRun)
	assert.Equal(t, 0, reqs.MaxSimilarRun)
}

func Test_longestWeakRun(t *testing.T) {
	tests := []struct {
		secret string
		want   int
	}{
		{"", 0},
		{"a", 1},
		{"xq", 1},
		{"123", 3},
		{"1234", 4},
		{"4321", 4},
		{"abcdef", 6},
		{"aaaa", 4},
		{"qwerty", 6},
		{"1213", 2},
		{"xk7q2m9z", 1},
	}
	for _, tt := range tests {
		t.Run(tt.secret, func(t *testing.T) {
			assert.Equal(t, tt.want, longestWeakRun([]rune(tt.secret)))
		})
	}
}

func Test_longestSharedRun_ReadsPrevBothWays(t *testing.T) {
	secret := []rune("xxdcbaxx")

	// "abcd" only overlaps once it is read backwards.
	assert.Equal(t, 1, longestCommonRun(secret, []rune("abcd")))
	assert.Equal(t, 4, longestSharedRun(secret, []rune("abcd")))

	// The forward direction still wins when it is the longer of the two.
	assert.Equal(t, 4, longestSharedRun([]rune("xxabcdxx"), []rune("abcd")))
}

func Test_reverseRunes(t *testing.T) {
	assert.Equal(t, []rune(""), reverseRunes([]rune("")))
	assert.Equal(t, []rune("a"), reverseRunes([]rune("a")))
	assert.Equal(t, []rune("cba"), reverseRunes([]rune("abc")))

	// The input must not be modified in place.
	in := []rune("abc")
	_ = reverseRunes(in)
	assert.Equal(t, []rune("abc"), in)
}

func Test_longestCommonRun(t *testing.T) {
	tests := []struct {
		name string
		a    string
		b    string
		want int
	}{
		{"empty a", "", "abc", 0},
		{"empty b", "abc", "", 0},
		{"no overlap", "abc", "xyz", 0},
		{"single character", "abc", "zzcz", 1},
		{"whole string", "abcd", "xxabcdxx", 4},
		{"case insensitive", "ABCD", "xxabcdxx", 4},
		{"longest of several", "abXcde", "cdeXab", 3},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, longestCommonRun([]rune(tt.a), []rune(tt.b)))
		})
	}
}
