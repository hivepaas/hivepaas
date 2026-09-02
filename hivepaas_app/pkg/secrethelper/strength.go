package secrethelper

import (
	"unicode"

	"github.com/tiendc/gofn"

	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
)

const (
	DefaultPasswordMinLen             = 8
	DefaultPasswordMaxLen             = 50
	DefaultPasswordRequiredLowercases = 1
	DefaultPasswordRequiredUppercases = 1
	DefaultPasswordRequiredDigits     = 1
	DefaultPasswordRequiredSymbols    = 1
)

var (
	specialCharset  = []rune("!@#$%^&*()_+-=[]{}|;':\",./<>?")
	mapSpecialChars = gofn.MapSliceToMap(specialCharset, func(k rune) (rune, struct{}) {
		return k, struct{}{}
	})
)

// ValidateStrength checks the password if it meets the strength requirements
// TODO: consider checking password must not contain first/last name, email
// TODO: consider checking password must not be the same as last 3 history passwords
func ValidateStrength(
	password string,
	minLen, maxLen int,
	requiredLowercases, requiredUppercases int,
	requiredDigits, requiredSymbols int,
) error {
	if minLen < 0 {
		minLen = DefaultPasswordMinLen
	}
	if maxLen < 0 {
		maxLen = DefaultPasswordMaxLen
	}
	if requiredLowercases < 0 {
		requiredLowercases = DefaultPasswordRequiredLowercases
	}
	if requiredUppercases < 0 {
		requiredUppercases = DefaultPasswordRequiredUppercases
	}
	if requiredDigits < 0 {
		requiredDigits = DefaultPasswordRequiredDigits
	}
	if requiredSymbols < 0 {
		requiredSymbols = DefaultPasswordRequiredSymbols
	}

	chars := []rune(password)
	if len(chars) < minLen || len(chars) > maxLen {
		return hperrors.Wrap(hperrors.ErrPasswordNotMeetRequirements).
			WithParams(getStrengthNotMeetErrParams(minLen, maxLen, requiredLowercases,
				requiredUppercases, requiredDigits, requiredSymbols)).
			WithMsgLog("incorrect length: %d", len(chars))
	}
	lowers := 0
	uppers := 0
	digits := 0
	specials := 0
	for _, r := range chars {
		switch {
		case unicode.IsDigit(r):
			digits++
		case gofn.MapContainKeys(mapSpecialChars, r):
			specials++
		case unicode.IsLower(r):
			lowers++
		default:
			uppers++
		}
	}
	if lowers < requiredLowercases || uppers < requiredUppercases ||
		digits < requiredDigits || specials < requiredSymbols {
		return hperrors.Wrap(hperrors.ErrPasswordNotMeetRequirements).
			WithParams(getStrengthNotMeetErrParams(minLen, maxLen, requiredLowercases,
				requiredUppercases, requiredDigits, requiredSymbols)).
			WithMsgLog("lowers %d, uppers %d, digits %d, specials %d", lowers, uppers, digits, specials)
	}
	return nil
}

func getStrengthNotMeetErrParams(
	minLen, maxLen int,
	requiredLowercases, requiredUppercases int,
	requiredDigits, requiredSymbols int,
) map[string]any {
	return map[string]any{
		"MinLen":   minLen,
		"MaxLen":   maxLen,
		"Lowers":   requiredLowercases,
		"Uppers":   requiredUppercases,
		"Digits":   requiredDigits,
		"Specials": requiredSymbols,
	}
}
