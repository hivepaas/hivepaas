package cryptoutil

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha256"
	"crypto/subtle"
	"encoding/base64"
	"encoding/binary"
	"io"
	"strings"
	"sync"

	"github.com/tiendc/gofn"
	"golang.org/x/crypto/argon2"

	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/reflectutil"
)

const (
	keyLen        = 32
	hashIteration = 1
	hashMemory    = 64 * 1024
	hashThread    = 2

	// keyCacheMaxEntries bounds the derived-key cache. Salts come from data this
	// app encrypted itself, so the working set is the number of stored secrets.
	keyCacheMaxEntries = 1024
)

// keyCache memoizes key derivation. Deriving a key costs 64 MiB and several
// milliseconds; the same salt is decrypted over and over - a repo webhook
// decrypts its secret on every delivery, on an unauthenticated route, before any
// signature is checked. Caching in process does not weaken the KDF, which exists
// to protect the app secret against an offline attack on stolen ciphertext.
var keyCache = struct {
	sync.RWMutex
	entries map[string][]byte
}{entries: make(map[string][]byte, keyCacheMaxEntries)}

func makeKey(secret, salt []byte) []byte {
	return argon2.IDKey(secret, salt, hashIteration, hashMemory, hashThread, keyLen)
}

// makeKeyCached is makeKey for decryption, where the same (secret, salt) pair
// recurs. Encryption draws a fresh salt every time, so it stays uncached.
func makeKeyCached(secret, salt []byte) []byte {
	cacheKey := keyCacheKey(secret, salt)

	keyCache.RLock()
	key, ok := keyCache.entries[cacheKey]
	keyCache.RUnlock()
	if ok {
		return key
	}

	key = makeKey(secret, salt)

	keyCache.Lock()
	// Dropping the whole cache when it grows too large keeps the bound simple;
	// it only ever costs a few re-derivations.
	if len(keyCache.entries) >= keyCacheMaxEntries {
		clear(keyCache.entries)
	}
	keyCache.entries[cacheKey] = key
	keyCache.Unlock()

	return key
}

// keyCacheKey binds an entry to both the salt and the app secret, so rotating the
// secret cannot serve a stale key. The length prefix keeps the two inputs from
// running together.
func keyCacheKey(secret, salt []byte) string {
	hash := sha256.New()
	_ = binary.Write(hash, binary.BigEndian, uint64(len(secret)))
	hash.Write(secret)
	hash.Write(salt)
	return string(hash.Sum(nil))
}

func Encrypt(plaintext, salt, secret []byte) ([]byte, error) {
	block, err := aes.NewCipher(makeKey(secret, salt))
	if err != nil {
		return nil, hperrors.Wrap(err)
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, hperrors.Wrap(err)
	}

	nonce := make([]byte, gcm.NonceSize())
	if _, err = io.ReadFull(rand.Reader, nonce); err != nil {
		return nil, hperrors.Wrap(err)
	}

	ciphertext := gcm.Seal(nonce, nonce, plaintext, nil)
	return ciphertext, nil
}

// EncryptBase64 this encrypts the input and returns a string in form: `hpsalt:<salt> <secret>`
func EncryptBase64(plaintext string, saltLen int, secret string) (ciphertext string, err error) {
	if saltLen <= 0 {
		return plaintext, nil
	}

	saltBytes := gofn.RandToken(saltLen)
	ciphertextBytes, err := Encrypt(reflectutil.UnsafeStrToBytes(plaintext), saltBytes,
		reflectutil.UnsafeStrToBytes(secret))
	if err != nil {
		return "", hperrors.Wrap(err)
	}
	ciphertext = base64.StdEncoding.EncodeToString(ciphertextBytes)
	salt := base64.StdEncoding.EncodeToString(saltBytes)
	return PackSecret(ciphertext, salt), nil
}

func Decrypt(ciphertext, salt, secret []byte) ([]byte, error) {
	block, err := aes.NewCipher(makeKeyCached(secret, salt))
	if err != nil {
		return nil, hperrors.Wrap(err)
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, hperrors.Wrap(err)
	}

	nonceSize := gcm.NonceSize()
	if len(ciphertext) < nonceSize {
		return nil, hperrors.NewArgumentInvalid("ciphertext").
			WithMsgLog("ciphertext too short")
	}

	nonce, ciphertext := ciphertext[:nonceSize], ciphertext[nonceSize:]
	plaintext, err := gcm.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return nil, hperrors.Wrap(err)
	}
	return plaintext, nil
}

// DecryptBase64 this decrypts the input in form: `hpsalt:<salt> <secret>`
func DecryptBase64(encryptedText, secret string) (plaintext string, err error) {
	ciphertext, salt := UnpackSecret(encryptedText)
	if salt == "" {
		return ciphertext, nil
	}

	ciphertextBytes, err := base64.StdEncoding.DecodeString(ciphertext)
	if err != nil {
		return "", hperrors.Wrap(err)
	}
	saltBytes, err := base64.StdEncoding.DecodeString(salt)
	if err != nil {
		return "", hperrors.Wrap(err)
	}

	plaintextBytes, err := Decrypt(ciphertextBytes, saltBytes, reflectutil.UnsafeStrToBytes(secret))
	if err != nil {
		return "", hperrors.Wrap(err)
	}
	plaintext = string(plaintextBytes)

	return plaintext, nil
}

func PackSecret(secret, salt string) string {
	return base.EncryptionSaltPrefix + salt + " " + secret
}

func UnpackSecret(secretText string) (secret string, salt string) {
	if !strings.HasPrefix(secretText, base.EncryptionSaltPrefix) {
		return secretText, ""
	}
	parts := strings.SplitN(secretText, " ", 2) //nolint:mnd
	if len(parts) != 2 {                        //nolint:mnd
		return secretText, ""
	}
	return parts[1], strings.TrimPrefix(parts[0], base.EncryptionSaltPrefix)
}

// SecureCompare compares two strings in constant time (O(N)) to prevent timing attacks.
func SecureCompare(a, b string) bool {
	return subtle.ConstantTimeCompare([]byte(a), []byte(b)) == 1
}

// SecureCompareBytes compares two byte slices in constant time (O(N)) to prevent timing attacks.
func SecureCompareBytes(a, b []byte) bool {
	return subtle.ConstantTimeCompare(a, b) == 1
}
