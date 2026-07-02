// Source copied from: https://github.com/foomo/htpasswd

package htpasswd

import (
	"errors"
	"os"

	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/tracerr"
)

var (
	ErrPasswordRequired     = errors.New("password is required")
	ErrUnsupportedAlgorithm = errors.New("unsupported hash algorithm")
)

// HashedPasswords name => hash
type HashedPasswords map[string]string

const (
	// PasswordSeparator separates passwords from hashes
	PasswordSeparator = ":"
	// LineSeparator separates password records
	LineSeparator = "\n"
)

const (
	fileMode = 0o644
)

// Bytes bytes representation
func (hp HashedPasswords) Bytes() (passwordBytes []byte) {
	passwordBytes = []byte{}
	for name, hash := range hp {
		passwordBytes = append(passwordBytes, []byte(name+PasswordSeparator+hash+LineSeparator)...)
	}
	return passwordBytes
}

// WriteToFile put them to a file will be overwritten or created
func (hp HashedPasswords) WriteToFile(file string) error {
	err := os.WriteFile(file, hp.Bytes(), fileMode)
	if err != nil {
		return tracerr.Wrap(err)
	}
	return nil
}

// SetPassword set a password for a user with a hashing algo
func (hp HashedPasswords) SetPassword(name, password string, hashAlgorithm HashAlgorithm) (err error) {
	if len(password) == 0 {
		return tracerr.Wrap(ErrPasswordRequired)
	}
	hash := ""
	switch hashAlgorithm {
	case HashBCrypt:
		hash, err = hashBcrypt(password)
	default:
		return tracerr.Wrap(ErrUnsupportedAlgorithm)
	}
	if err != nil {
		return tracerr.Wrap(err)
	}
	hp[name] = hash
	return nil
}
