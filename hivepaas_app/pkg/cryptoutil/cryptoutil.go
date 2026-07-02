package cryptoutil

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"io"
	"strings"

	"github.com/tiendc/gofn"
	"golang.org/x/crypto/argon2"

	"github.com/hivepaas/hivepaas/hivepaas_app/apperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/reflectutil"
)

const (
	keyLen        = 32
	hashIteration = 1
	hashMemory    = 64 * 1024
	hashThread    = 2
)

func makeKey(secret, salt []byte) []byte {
	return argon2.IDKey(secret, salt, hashIteration, hashMemory, hashThread, keyLen)
}

func Encrypt(plaintext, salt, secret []byte) ([]byte, error) {
	block, err := aes.NewCipher(makeKey(secret, salt))
	if err != nil {
		return nil, apperrors.New(err)
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, apperrors.New(err)
	}

	nonce := make([]byte, gcm.NonceSize())
	if _, err = io.ReadFull(rand.Reader, nonce); err != nil {
		return nil, apperrors.New(err)
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
		return "", apperrors.New(err)
	}
	ciphertext = base64.StdEncoding.EncodeToString(ciphertextBytes)
	salt := base64.StdEncoding.EncodeToString(saltBytes)
	return PackSecret(ciphertext, salt), nil
}

func Decrypt(ciphertext, salt, secret []byte) ([]byte, error) {
	block, err := aes.NewCipher(makeKey(secret, salt))
	if err != nil {
		return nil, apperrors.New(err)
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, apperrors.New(err)
	}

	nonceSize := gcm.NonceSize()
	if len(ciphertext) < nonceSize {
		return nil, apperrors.NewArgumentInvalid("ciphertext").
			WithMsgLog("ciphertext too short")
	}

	nonce, ciphertext := ciphertext[:nonceSize], ciphertext[nonceSize:]
	plaintext, err := gcm.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return nil, apperrors.New(err)
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
		return "", apperrors.New(err)
	}
	saltBytes, err := base64.StdEncoding.DecodeString(salt)
	if err != nil {
		return "", apperrors.New(err)
	}

	plaintextBytes, err := Decrypt(ciphertextBytes, saltBytes, reflectutil.UnsafeStrToBytes(secret))
	if err != nil {
		return "", apperrors.New(err)
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
