package randtoken

import (
	"bytes"
	"crypto/rand"
	"encoding/hex"
	"fmt"

	"golang.org/x/crypto/argon2"
)

const (
	hashingMemory  = 64 * 1024 // 64MB
	hashingThreads = 4
)

func Hash(token []byte, saltLen, hashingKeyLen, hashingIteration uint32) (hash []byte, salt []byte, err error) {
	salt = make([]byte, saltLen)
	if _, err = rand.Read(salt); err != nil {
		return nil, nil, fmt.Errorf("failed to generate salt: %w", err)
	}

	// Hash the token using Argon2 with recommended configuration
	hash = argon2.IDKey(token, salt, hashingIteration, hashingMemory, hashingThreads, hashingKeyLen)
	return hash, salt, nil
}

func HashAsHex(token string, saltLen, hashingKeyLen, hashingIteration uint32) (string, string, error) {
	tokenBytes, err := hex.DecodeString(token)
	if err != nil {
		return "", "", fmt.Errorf("failed to decode token as hex: %w", err)
	}
	hash, salt, err := Hash(tokenBytes, saltLen, hashingKeyLen, hashingIteration)
	if err != nil {
		return "", "", fmt.Errorf("failed to hash token: %w", err)
	}
	return fmt.Sprintf("%x", hash), fmt.Sprintf("%x", salt), nil
}

func VerifyHash(token, hash, salt []byte, hashingKeyLen, hashingIteration uint32) bool {
	if len(token) == 0 || len(hash) == 0 {
		return false
	}
	thisHash := argon2.IDKey(token, salt, hashingIteration, hashingMemory, hashingThreads, hashingKeyLen)
	return bytes.Equal(thisHash, hash)
}

func VerifyHashHex(token, hash, salt string, hashingKeyLen, hashingIteration uint32) bool {
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
	return VerifyHash(tokenBytes, hashBytes, saltBytes, hashingKeyLen, hashingIteration)
}
