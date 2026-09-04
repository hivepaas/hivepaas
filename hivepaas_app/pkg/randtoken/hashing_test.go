package randtoken

import (
	"crypto/sha256"
	"encoding/hex"
	"testing"

	"github.com/stretchr/testify/assert"
)

const testSaltLen = 16

func TestHashAndVerify(t *testing.T) {
	token := []byte("secret-token")

	hash, salt, err := HashNew(token, testSaltLen)
	assert.NoError(t, err)
	assert.Len(t, salt, testSaltLen)
	assert.Len(t, hash, sha256.Size)

	assert.True(t, VerifyHash(token, hash, salt))

	// Wrong token
	assert.False(t, VerifyHash([]byte("wrong-token"), hash, salt))

	// Wrong salt
	wrongSalt := make([]byte, testSaltLen)
	copy(wrongSalt, salt)
	wrongSalt[0] ^= 0xFF
	assert.False(t, VerifyHash(token, hash, wrongSalt))

	// Empty inputs
	assert.False(t, VerifyHash(nil, hash, salt))
	assert.False(t, VerifyHash(token, nil, salt))
}

// Two tokens hashed separately must not collide through their salts, and the same
// token must hash differently under different salts.
func TestHashNewDrawsAFreshSalt(t *testing.T) {
	token := []byte("secret-token")

	firstHash, firstSalt, err := HashNew(token, testSaltLen)
	assert.NoError(t, err)
	secondHash, secondSalt, err := HashNew(token, testSaltLen)
	assert.NoError(t, err)

	assert.NotEqual(t, firstSalt, secondSalt)
	assert.NotEqual(t, firstHash, secondHash)
}

// The salt has to take part in the hash, otherwise it is decoration.
func TestHashDependsOnSalt(t *testing.T) {
	token := []byte("secret-token")
	assert.NotEqual(t, Hash(token, []byte("salt-a")), Hash(token, []byte("salt-b")))
}

func TestHashAndVerifyHex(t *testing.T) {
	tokenHex := hex.EncodeToString([]byte("secret-token"))

	hashHex, saltHex, err := HashNewAsHex(tokenHex, testSaltLen)
	assert.NoError(t, err)
	assert.NotEmpty(t, hashHex)
	assert.NotEmpty(t, saltHex)

	assert.True(t, VerifyHashHex(tokenHex, hashHex, saltHex))

	wrongTokenHex := hex.EncodeToString([]byte("wrong-token"))
	assert.False(t, VerifyHashHex(wrongTokenHex, hashHex, saltHex))

	// Malformed hex must be rejected, not panic
	assert.False(t, VerifyHashHex("invalid-hex", hashHex, saltHex))
	assert.False(t, VerifyHashHex(tokenHex, "invalid-hex", saltHex))
	assert.False(t, VerifyHashHex(tokenHex, hashHex, "invalid-hex"))
}

func TestHashNewAsHexInvalidInput(t *testing.T) {
	_, _, err := HashNewAsHex("invalid-hex", testSaltLen)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to decode token as hex")
}
