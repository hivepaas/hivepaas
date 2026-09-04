// Package randtoken hashes high-entropy random tokens - API key secrets and the
// like - for storage.
//
// It deliberately uses a plain salted SHA-256 rather than a password hash such as
// argon2. Memory-hard hashing exists to make brute force expensive against
// low-entropy, human-chosen inputs; a token drawn from crypto/rand has far too
// much entropy for brute force to be a threat regardless of how fast the hash is.
// Paying for argon2 here would only buy 64 MiB and ~10ms on every request that
// authenticates with an API key.
//
// Passwords are a different matter and must keep using argon2 - see
// userservice.createPasswordHash.
package randtoken

import (
	"crypto/rand"
	"crypto/sha256"
	"crypto/subtle"
	"encoding/hex"
	"fmt"
)

// Hash returns the SHA-256 of salt+token.
func Hash(token, salt []byte) []byte {
	hash := sha256.New()
	hash.Write(salt)
	hash.Write(token)
	return hash.Sum(nil)
}

// HashNew draws a fresh salt and hashes the token with it.
func HashNew(token []byte, saltLen int) (hash []byte, salt []byte, err error) {
	salt = make([]byte, saltLen)
	if _, err = rand.Read(salt); err != nil {
		return nil, nil, fmt.Errorf("failed to generate salt: %w", err)
	}
	return Hash(token, salt), salt, nil
}

// HashNewAsHex is HashNew over a hex-encoded token, returning hex.
func HashNewAsHex(token string, saltLen int) (hashHex string, saltHex string, err error) {
	tokenBytes, err := hex.DecodeString(token)
	if err != nil {
		return "", "", fmt.Errorf("failed to decode token as hex: %w", err)
	}
	hash, salt, err := HashNew(tokenBytes, saltLen)
	if err != nil {
		return "", "", fmt.Errorf("failed to hash token: %w", err)
	}
	return hex.EncodeToString(hash), hex.EncodeToString(salt), nil
}

// VerifyHash reports whether the token hashes to the stored value. The comparison
// is constant time: the caller controls the token, so it controls the computed
// hash, and a leaky comparison would expose the stored one byte by byte.
func VerifyHash(token, hash, salt []byte) bool {
	if len(token) == 0 || len(hash) == 0 {
		return false
	}
	return subtle.ConstantTimeCompare(Hash(token, salt), hash) == 1
}

// VerifyHashHex is VerifyHash over hex-encoded inputs.
func VerifyHashHex(token, hash, salt string) bool {
	tokenBytes, err := hex.DecodeString(token)
	if err != nil {
		return false
	}
	hashBytes, err := hex.DecodeString(hash)
	if err != nil {
		return false
	}
	saltBytes, err := hex.DecodeString(salt)
	if err != nil {
		return false
	}
	return VerifyHash(tokenBytes, hashBytes, saltBytes)
}
