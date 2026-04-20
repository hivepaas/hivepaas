package cryptoutil

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/localpaas/localpaas/localpaas_app/base"
)

func TestPackUnpackSecret(t *testing.T) {
	secret := "ciphertext"
	salt := "randomsalt"
	packed := PackSecret(secret, salt)
	// Ensure prefix is present
	assert.True(t, strings.HasPrefix(packed, base.EncryptionSaltPrefix))
	// Unpack should retrieve original values
	gotSecret, gotSalt := UnpackSecret(packed)
	assert.Equal(t, secret, gotSecret)
	assert.Equal(t, salt, gotSalt)
}

func TestEncryptDecryptRoundTrip(t *testing.T) {
	plaintext := []byte("hello world")
	salt := []byte("fixedsalthere1234567890abcdef")    // 32 bytes for example
	secret := []byte("supersecretkeysupersecretkey12") // 32 bytes

	// Encrypt
	ct, err := Encrypt(plaintext, salt, secret)
	assert.NoError(t, err)
	// Decrypt
	pt, err := Decrypt(ct, salt, secret)
	assert.NoError(t, err)
	assert.Equal(t, plaintext, pt)
}

func TestEncryptBase64DecryptBase64(t *testing.T) {
	// Use a small salt length to keep test fast
	plaintext := "sample text for encryption"
	secret := "myverystrongsecret"
	encrypted, err := EncryptBase64(plaintext, 8, secret)
	assert.NoError(t, err)

	// Decrypt back
	decrypted, err := DecryptBase64(encrypted, secret)
	assert.NoError(t, err)
	assert.Equal(t, plaintext, decrypted)
}

func TestEncryptBase64NoSalt(t *testing.T) {
	plaintext := "nosalttext"
	secret := "secret"
	// saltLen <= 0 should return plaintext unchanged
	got, err := EncryptBase64(plaintext, 0, secret)
	assert.NoError(t, err)
	assert.Equal(t, plaintext, got)
}
