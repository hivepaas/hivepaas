package secrethelper

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
)

func Test_ValidateStrength(t *testing.T) {
	errFail := hperrors.ErrPasswordNotMeetRequirements
	tests := []struct {
		name     string
		password string
		wantErr  error
	}{
		{"too short", "Ab1!123", errFail},
		{"too long", "A" + strings.Repeat("a", DefaultPasswordMaxLen-2) + "1!", errFail},
		{"no lowercase", "A1!A1!", errFail},
		{"no uppercase", "a1!a1!", errFail},
		{"no digit", "Aa!Aa!", errFail},
		{"no symbol", "Aa1Aa1", errFail},
		{"valid password", "Aa1!Aa1!", nil},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateStrength(tt.password, DefaultPasswordMinLen, DefaultPasswordMaxLen,
				DefaultPasswordRequiredLowercases, DefaultPasswordRequiredUppercases,
				DefaultPasswordRequiredDigits, DefaultPasswordRequiredSymbols)
			if tt.wantErr != nil {
				assert.ErrorIs(t, err, tt.wantErr)
			} else {
				assert.Nil(t, err)
			}
		})
	}
}
