package entity

import (
	"strings"

	"github.com/tiendc/gofn"

	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/datakey"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/reflectutil"
)

const (
	defaultSaltLen = 10
)

type EncryptedField struct {
	encrypted string
	decrypted string
}

func (s *EncryptedField) MarshalJSON() (res []byte, err error) {
	// Encryption is not optional. This used to fall back to writing s.encrypted
	// when no key was configured, which for a field holding a plaintext value
	// wrote an empty string and lost it without any error. Refuse instead:
	// startup installs a data key, so reaching this is a bug.
	encrypted, err := s.encrypt()
	if err != nil {
		return nil, hperrors.Wrap(err)
	}
	return reflectutil.UnsafeStrToBytes(gofn.StringWrap(encrypted, "\"")), nil
}

func (s *EncryptedField) UnmarshalJSON(data []byte) error {
	s.Set(gofn.StringUnwrap(reflectutil.UnsafeBytesToStr(data), "\""))
	return nil
}

func (s *EncryptedField) String() string {
	if s.decrypted != "" {
		return s.decrypted
	}
	return s.encrypted
}

func (s *EncryptedField) IsEmpty() bool {
	return s.encrypted == "" && s.decrypted == ""
}

func (s *EncryptedField) IsEncrypted() bool {
	return s.encrypted != "" && s.decrypted == ""
}

func (s *EncryptedField) GetPlain() (string, error) {
	decrypted, err := s.decrypt()
	if err != nil {
		return "", hperrors.Wrap(err)
	}
	return decrypted, nil
}

func (s *EncryptedField) GetEncrypted() (string, error) {
	encrypted, err := s.encrypt()
	if err != nil {
		return "", hperrors.Wrap(err)
	}
	return encrypted, nil
}

func (s *EncryptedField) Set(value string) {
	if isEncryptedValue(value) {
		s.encrypted = value
		s.decrypted = ""
	} else {
		s.decrypted = value
		s.encrypted = ""
	}
}

func (s *EncryptedField) Equal(enc *EncryptedField) (bool, error) {
	v1, err := s.GetPlain()
	if err != nil {
		return false, hperrors.Wrap(err)
	}
	v2, err := enc.GetPlain()
	if err != nil {
		return false, hperrors.Wrap(err)
	}
	return v1 == v2, nil
}

func (s *EncryptedField) encrypt() (string, error) {
	if s.encrypted != "" {
		return s.encrypted, nil
	}
	if s.decrypted == "" {
		return "", nil // nothing to encrypt, the field is unset
	}
	key := datakey.Active()
	if key == nil {
		return "", hperrors.NewMissing("Data encryption key")
	}
	encrypted, err := key.Seal(s.decrypted)
	if err != nil {
		return "", hperrors.Wrap(err)
	}
	s.encrypted = encrypted
	return encrypted, nil
}

func (s *EncryptedField) decrypt() (string, error) {
	if s.decrypted != "" {
		return s.decrypted, nil
	}
	if s.encrypted == "" {
		return "", nil // nothing to decrypt, the field is unset
	}
	key := datakey.Active()
	if key == nil {
		return "", hperrors.NewMissing("Data encryption key")
	}
	decrypted, err := key.Open(s.encrypted)
	if err != nil {
		return "", hperrors.Wrap(err)
	}
	s.decrypted = decrypted
	return decrypted, nil
}

// isEncryptedValue reports whether a stored value is ciphertext rather than a
// plaintext one that still has to be encrypted.
func isEncryptedValue(value string) bool {
	for _, prefix := range base.AllEncryptionPrefixes {
		if strings.HasPrefix(value, prefix) {
			return true
		}
	}
	return false
}

// Reencrypt drops the cached ciphertext so the next marshal encrypts the value
// again with whatever key is current.
//
// Marshaling normally reuses the ciphertext the value was loaded with, which is
// exactly what re-saving an unrelated setting should do. Rotating the app secret
// is the one case that has to bypass it.
func (s *EncryptedField) Reencrypt() error {
	if s.IsEmpty() {
		return nil
	}
	plain, err := s.GetPlain()
	if err != nil {
		return hperrors.Wrap(err)
	}
	s.Set(plain)
	return nil
}

func NewEncryptedField(value string) EncryptedField {
	resp := EncryptedField{}
	resp.Set(value)
	return resp
}
