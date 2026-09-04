package cryptoutil

import (
	"bytes"
	"testing"
)

func TestMakeKeyCachedMatchesUncached(t *testing.T) {
	secret := []byte("app-secret")
	salt := []byte("some-salt")

	want := makeKey(secret, salt)
	if got := makeKeyCached(secret, salt); !bytes.Equal(got, want) {
		t.Fatal("cached key differs from the derived one")
	}
	// Second call comes from the cache and must still match.
	if got := makeKeyCached(secret, salt); !bytes.Equal(got, want) {
		t.Fatal("cached key differs on the second call")
	}
}

// Rotating the app secret must not serve the key derived from the old one.
func TestMakeKeyCachedIsBoundToSecretAndSalt(t *testing.T) {
	salt := []byte("same-salt")

	first := makeKeyCached([]byte("secret-a"), salt)
	second := makeKeyCached([]byte("secret-b"), salt)
	if bytes.Equal(first, second) {
		t.Error("two different secrets produced the same key")
	}

	secret := []byte("same-secret")
	if bytes.Equal(makeKeyCached(secret, []byte("salt-a")), makeKeyCached(secret, []byte("salt-b"))) {
		t.Error("two different salts produced the same key")
	}
}

// The cache key must not let a (secret, salt) pair collide with a different split
// of the same bytes.
func TestKeyCacheKeyIsUnambiguous(t *testing.T) {
	if keyCacheKey([]byte("ab"), []byte("c")) == keyCacheKey([]byte("a"), []byte("bc")) {
		t.Error("cache key runs the secret and the salt together")
	}
}

func TestKeyCacheStaysBounded(t *testing.T) {
	// Fill the cache with placeholders rather than real derivations: deriving a
	// thousand keys would cost a thousand 64 MiB argon2 runs.
	keyCache.Lock()
	clear(keyCache.entries)
	for i := range keyCacheMaxEntries {
		keyCache.entries[string(rune(i))] = nil
	}
	keyCache.Unlock()

	makeKeyCached([]byte("app-secret"), []byte("one-more-salt"))

	keyCache.RLock()
	size := len(keyCache.entries)
	keyCache.RUnlock()

	if size > keyCacheMaxEntries {
		t.Errorf("cache grew past its bound: %d entries", size)
	}
}

func TestEncryptDecryptRoundTripWithCache(t *testing.T) {
	const secret = "app-secret"

	encrypted, err := EncryptBase64("s3cr3t-value", defaultTestSaltLen, secret)
	if err != nil {
		t.Fatalf("EncryptBase64: %v", err)
	}
	for range 3 { // repeated decrypts go through the cache
		plain, err := DecryptBase64(encrypted, secret)
		if err != nil {
			t.Fatalf("DecryptBase64: %v", err)
		}
		if plain != "s3cr3t-value" {
			t.Fatalf("round trip returned %q", plain)
		}
	}

	if _, err := DecryptBase64(encrypted, "wrong-secret"); err == nil {
		t.Error("decrypting with the wrong secret should fail")
	}
}

const defaultTestSaltLen = 10
