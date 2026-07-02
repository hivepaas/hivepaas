package htpasswd

import (
	"golang.org/x/crypto/bcrypt"

	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/tracerr"
)

// HashAlgorithm enum for hashing algorithms
type HashAlgorithm string

const (
	// HashBCrypt bcrypt - recommended
	HashBCrypt = "bcrypt"
)

func HashPassword(password string) (string, error) {
	return hashBcrypt(password)
}

func hashBcrypt(password string) (string, error) {
	passwordBytes, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", tracerr.Wrap(err)
	}
	return string(passwordBytes), nil
}
