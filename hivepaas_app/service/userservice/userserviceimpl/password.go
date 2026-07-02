package userserviceimpl

import (
	"bytes"
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"strings"
	"unicode"

	"github.com/tiendc/gofn"
	"golang.org/x/crypto/argon2"

	"github.com/hivepaas/hivepaas/hivepaas_app/apperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
)

const (
	saltLength       = 10
	hashingIteration = 1
	hashingMemory    = 64 * 1024 // 64MB
	hashingThreads   = 4
	hashingKeyLength = 32
)

const (
	passwordMinLen             = 8
	passwordMaxLen             = 32
	passwordMustHaveLowercases = 1
	passwordMustHaveUppercases = 1
	passwordMustHaveDigits     = 1
	passwordMustHaveSymbols    = 1
)

var (
	errNotMeetRequirementParams = map[string]any{
		"MinLen":   passwordMinLen,
		"MaxLen":   passwordMaxLen,
		"Lowers":   passwordMustHaveLowercases,
		"Uppers":   passwordMustHaveUppercases,
		"Digits":   passwordMustHaveDigits,
		"Specials": passwordMustHaveSymbols,
	}

	specialCharset  = []rune("!@#$%^&*()_+-=[]{}|;':\",./<>?")
	mapSpecialChars = gofn.MapSliceToMap(specialCharset, func(k rune) (rune, struct{}) {
		return k, struct{}{}
	})
)

// ChangePassword updates user password with the new one.
// Action is rejected if `currPassword` does not match.
// Checking for current password is skipped if pass empty string.
func (s *service) ChangePassword(user *entity.User, newPassword, currPassword string) (err error) {
	if currPassword != "" {
		if err := s.VerifyPassword(user, currPassword); err != nil {
			return err
		}
	}

	// Verify password strength
	if err := s.CheckPasswordStrength(newPassword); err != nil {
		return err
	}

	user.Password, err = s.createPasswordHash(newPassword)
	if err != nil {
		return fmt.Errorf("failed to generate password hash: %w", err)
	}
	return nil
}

func (s *service) createPasswordHash(password string) (string, error) {
	salt := make([]byte, saltLength)
	if _, err := rand.Read(salt); err != nil {
		return "", fmt.Errorf("failed to generate salt: %w", err)
	}

	// Hash the password using Argon2 with recommended configuration
	hashedPass := argon2.IDKey([]byte(password), salt, hashingIteration, hashingMemory,
		hashingThreads, hashingKeyLength)

	return base64.StdEncoding.EncodeToString(salt) + " " +
		base64.StdEncoding.EncodeToString(hashedPass), nil
}

// VerifyPassword verifies the password matching the user data
func (s *service) VerifyPassword(user *entity.User, password string) error {
	// We don't allow empty password
	if password == "" || len(user.Password) == 0 {
		return apperrors.New(apperrors.ErrPasswordMismatched)
	}

	b64Salt, b64Pass, _ := strings.Cut(user.Password, " ")
	saltBytes, _ := base64.StdEncoding.DecodeString(b64Salt)
	passBytes, _ := base64.StdEncoding.DecodeString(b64Pass)

	passHash := argon2.IDKey([]byte(password), saltBytes, hashingIteration, hashingMemory,
		hashingThreads, hashingKeyLength)
	if !bytes.Equal(passHash, passBytes) {
		return apperrors.New(apperrors.ErrPasswordMismatched)
	}
	return nil
}

// CheckPasswordStrength checks the password if it meets the strength requirements
// TODO: consider checking password must not contain first/last name, email
// TODO: consider checking password must not be the same as last 3 history passwords
func (s *service) CheckPasswordStrength(password string) error {
	chars := []rune(password)
	if len(chars) < passwordMinLen || len(chars) > passwordMaxLen {
		return apperrors.New(apperrors.ErrPasswordNotMeetRequirements).
			WithParams(errNotMeetRequirementParams).
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
	if lowers < passwordMustHaveLowercases || uppers < passwordMustHaveUppercases ||
		digits < passwordMustHaveDigits || specials < passwordMustHaveSymbols {
		return apperrors.New(apperrors.ErrPasswordNotMeetRequirements).
			WithParams(errNotMeetRequirementParams).
			WithMsgLog("lowers %d, uppers %d, digits %d, specials %d", lowers, uppers, digits, specials)
	}
	return nil
}
