package randtoken

import (
	"encoding/hex"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestHashAndVerify(t *testing.T) {
	token := []byte("secret-token")
	saltLen := uint32(16)
	keyLen := uint32(32)
	iterations := uint32(1)

	hash, salt, err := Hash(token, saltLen, keyLen, iterations)
	assert.NoError(t, err)
	assert.NotNil(t, salt)
	assert.Len(t, salt, int(saltLen))
	assert.NotNil(t, hash)
	assert.Len(t, hash, int(keyLen))

	// Verify success
	assert.True(t, VerifyHash(token, hash, salt, keyLen, iterations))

	// Verify failure - wrong token
	assert.False(t, VerifyHash([]byte("wrong-token"), hash, salt, keyLen, iterations))

	// Verify failure - wrong salt
	wrongSalt := make([]byte, saltLen)
	copy(wrongSalt, salt)
	wrongSalt[0] ^= 0xFF
	assert.False(t, VerifyHash(token, hash, wrongSalt, keyLen, iterations))

	// Verify failure - empty inputs
	assert.False(t, VerifyHash(nil, hash, salt, keyLen, iterations))
	assert.False(t, VerifyHash(token, nil, salt, keyLen, iterations))
}

func TestHashAndVerifyHex(t *testing.T) {
	tokenBytes := []byte("secret-token")
	tokenHex := hex.EncodeToString(tokenBytes)
	saltLen := uint32(16)
	keyLen := uint32(32)
	iterations := uint32(1)

	hashHex, saltHex, err := HashAsHex(tokenHex, saltLen, keyLen, iterations)
	assert.NoError(t, err)
	assert.NotEmpty(t, hashHex)
	assert.NotEmpty(t, saltHex)

	// Verify success
	assert.True(t, VerifyHashHex(tokenHex, hashHex, saltHex, keyLen, iterations))

	// Verify failure - wrong token
	wrongTokenHex := hex.EncodeToString([]byte("wrong-token"))
	assert.False(t, VerifyHashHex(wrongTokenHex, hashHex, saltHex, keyLen, iterations))

	// Verify failure - invalid hex inputs
	assert.False(t, VerifyHashHex("invalid-hex", hashHex, saltHex, keyLen, iterations))
	assert.False(t, VerifyHashHex(tokenHex, "invalid-hex", saltHex, keyLen, iterations))
	assert.False(t, VerifyHashHex(tokenHex, hashHex, "invalid-hex", keyLen, iterations))
}

func TestHashAsHex_InvalidInput(t *testing.T) {
	_, _, err := HashAsHex("invalid-hex", 16, 32, 1)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to decode token as hex")
}
