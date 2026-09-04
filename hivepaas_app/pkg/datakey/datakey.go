// Package datakey holds the key that encrypts stored secrets.
//
// Encryption is split in two levels. A data encryption key (DEK) - 32 random
// bytes - encrypts the values themselves, and is kept in the database wrapped by
// the app secret, which acts as the key encryption key (KEK).
//
// The split buys two things. Changing the app secret only re-wraps the DEK, one
// row, instead of re-encrypting every stored value. And because the DEK is
// random rather than operator-chosen, sealing a value needs no key derivation at
// all: deriving a key per value from the app secret cost 17ms and 64MiB each,
// which a process reading a few hundred settings paid on every start.
package datakey

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"io"
	"strings"
	"sync/atomic"

	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/cryptoutil"
)

// KeyLen is the DEK size in bytes, an AES-256 key.
const KeyLen = 32

// wrapSaltLen is the salt length used when wrapping the DEK with the app secret.
const wrapSaltLen = 10

var (
	ErrNoActiveKey = fmt.Errorf("no active data encryption key")
	ErrKeyInvalid  = fmt.Errorf("invalid data encryption key")
)

// Key is a data encryption key held in memory.
type Key struct {
	aead cipher.AEAD
	raw  []byte
}

// active is the key every seal and open goes through. It is process state, set
// once at startup, the same way config.Current is.
var active atomic.Pointer[Key]

// SetActive installs the key the app encrypts with.
func SetActive(key *Key) {
	active.Store(key)
}

// Active returns the key in use, or nil before startup has installed one.
func Active() *Key {
	return active.Load()
}

// New builds a key from raw bytes.
func New(raw []byte) (*Key, error) {
	if len(raw) != KeyLen {
		return nil, fmt.Errorf("%w: got %d bytes, want %d", ErrKeyInvalid, len(raw), KeyLen)
	}

	block, err := aes.NewCipher(raw)
	if err != nil {
		return nil, fmt.Errorf("%w: %w", ErrKeyInvalid, err)
	}
	aead, err := cipher.NewGCM(block)
	if err != nil {
		return nil, fmt.Errorf("%w: %w", ErrKeyInvalid, err)
	}
	return &Key{aead: aead, raw: raw}, nil
}

// Generate draws a new key.
func Generate() (*Key, error) {
	raw := make([]byte, KeyLen)
	if _, err := rand.Read(raw); err != nil {
		return nil, fmt.Errorf("failed to generate a data encryption key: %w", err)
	}
	return New(raw)
}

// Seal encrypts a value for storage. The result carries a prefix so a stored
// value can be told apart from a plaintext one.
func (k *Key) Seal(plaintext string) (string, error) {
	nonce := make([]byte, k.aead.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return "", fmt.Errorf("failed to generate a nonce: %w", err)
	}

	sealed := k.aead.Seal(nonce, nonce, []byte(plaintext), nil)
	return base.EncryptionKeyPrefix + base64.StdEncoding.EncodeToString(sealed), nil
}

// Open decrypts a stored value.
func (k *Key) Open(ciphertext string) (string, error) {
	raw, err := base64.StdEncoding.DecodeString(strings.TrimPrefix(ciphertext, base.EncryptionKeyPrefix))
	if err != nil {
		return "", fmt.Errorf("failed to decode the stored value: %w", err)
	}
	if len(raw) < k.aead.NonceSize() {
		return "", fmt.Errorf("%w: stored value is too short", ErrKeyInvalid)
	}

	nonce, sealed := raw[:k.aead.NonceSize()], raw[k.aead.NonceSize():]
	plaintext, err := k.aead.Open(nil, nonce, sealed, nil)
	if err != nil {
		return "", fmt.Errorf("failed to decrypt the stored value: %w", err)
	}
	return string(plaintext), nil
}

// Wrap encrypts the key itself with the app secret, for storage in the database.
//
// This one keeps the slow key derivation: the app secret is chosen by whoever
// deploys the app and may be weak, and wrapping happens once at startup rather
// than once per value.
func (k *Key) Wrap(appSecret string) (string, error) {
	wrapped, err := cryptoutil.EncryptBase64(
		base64.StdEncoding.EncodeToString(k.raw), wrapSaltLen, appSecret)
	if err != nil {
		return "", fmt.Errorf("failed to wrap the data encryption key: %w", err)
	}
	return wrapped, nil
}

// Unwrap recovers a key wrapped with Wrap.
func Unwrap(wrapped, appSecret string) (*Key, error) {
	encoded, err := cryptoutil.DecryptBase64(wrapped, appSecret)
	if err != nil {
		return nil, fmt.Errorf("failed to unwrap the data encryption key: %w", err)
	}
	raw, err := base64.StdEncoding.DecodeString(encoded)
	if err != nil {
		return nil, fmt.Errorf("%w: %w", ErrKeyInvalid, err)
	}
	return New(raw)
}
