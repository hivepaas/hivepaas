package userserviceimpl

import (
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"strings"

	"golang.org/x/crypto/argon2"

	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/cryptoutil"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/secrethelper"
)

const (
	saltLength       = 10
	hashingIteration = 1
	hashingMemory    = 64 * 1024 // 64MB
	hashingThreads   = 4
	hashingKeyLength = 32
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
		return hperrors.Wrap(hperrors.ErrPasswordMismatched)
	}

	b64Salt, b64Pass, _ := strings.Cut(user.Password, " ")
	saltBytes, _ := base64.StdEncoding.DecodeString(b64Salt)
	passBytes, _ := base64.StdEncoding.DecodeString(b64Pass)

	// Argon2 is the right choice here and must stay: passwords are human-chosen and
	// low entropy, which is exactly what memory-hard hashing defends against.
	passHash := argon2.IDKey([]byte(password), saltBytes, hashingIteration, hashingMemory,
		hashingThreads, hashingKeyLength)
	// Constant time: the caller controls the password, so it controls passHash, and
	// a leaky comparison would expose the stored hash byte by byte.
	if !cryptoutil.SecureCompareBytes(passHash, passBytes) {
		return hperrors.Wrap(hperrors.ErrPasswordMismatched)
	}
	return nil
}

func (s *service) CheckPasswordStrength(password string) error {
	err := secrethelper.ValidateStrength(password, -1, -1, -1, -1, -1, -1)
	if err != nil {
		return hperrors.Wrap(err)
	}
	return nil
}
