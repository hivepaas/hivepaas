package htpasswd

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"golang.org/x/crypto/bcrypt"
)

func TestHashPassword(t *testing.T) {
	hash, err := HashPassword("mysecret")
	assert.NoError(t, err)
	assert.NotEmpty(t, hash)

	// Verify it's a valid bcrypt hash
	err = bcrypt.CompareHashAndPassword([]byte(hash), []byte("mysecret"))
	assert.NoError(t, err)

	err = bcrypt.CompareHashAndPassword([]byte(hash), []byte("wrong"))
	assert.Error(t, err)
}

func TestHashedPasswords_Bytes(t *testing.T) {
	hp := make(HashedPasswords)
	hp["user1"] = "hash1"
	hp["user2"] = "hash2"

	b := hp.Bytes()
	str := string(b)

	// Maps are unordered, so check both combinations
	assert.True(t, strings.Contains(str, "user1:hash1\n"))
	assert.True(t, strings.Contains(str, "user2:hash2\n"))
	assert.Equal(t, len("user1:hash1\nuser2:hash2\n"), len(str))
}

func TestHashedPasswords_WriteToFile(t *testing.T) {
	tmpDir := t.TempDir()
	filePath := filepath.Join(tmpDir, "htpasswd")

	hp := make(HashedPasswords)
	hp["admin"] = "secret_hash"

	err := hp.WriteToFile(filePath)
	assert.NoError(t, err)

	content, err := os.ReadFile(filePath)
	assert.NoError(t, err)
	assert.Equal(t, "admin:secret_hash\n", string(content))
}

func TestHashedPasswords_SetPassword(t *testing.T) {
	t.Run("Valid BCrypt", func(t *testing.T) {
		hp := make(HashedPasswords)
		err := hp.SetPassword("user", "mypass", HashBCrypt)
		assert.NoError(t, err)

		hash, exists := hp["user"]
		assert.True(t, exists)
		assert.NotEmpty(t, hash)

		// Verify it generated a valid bcrypt hash
		err = bcrypt.CompareHashAndPassword([]byte(hash), []byte("mypass"))
		assert.NoError(t, err)
	})

	t.Run("Empty Password", func(t *testing.T) {
		hp := make(HashedPasswords)
		err := hp.SetPassword("user", "", HashBCrypt)
		assert.Error(t, err)
		assert.True(t, errors.Is(err, ErrPasswordRequired))
	})

	t.Run("Unsupported Algorithm", func(t *testing.T) {
		hp := make(HashedPasswords)
		err := hp.SetPassword("user", "mypass", HashAlgorithm("md5"))
		assert.Error(t, err)
		assert.True(t, errors.Is(err, ErrUnsupportedAlgorithm))
	})
}
