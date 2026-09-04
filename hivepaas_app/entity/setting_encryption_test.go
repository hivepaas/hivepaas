package entity

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/datakey"
)

type encryptedHolder struct {
	Secret EncryptedField `json:"secret"`
}

// useDataKey installs a fresh data key, the way startup does.
func useDataKey(t *testing.T) *datakey.Key {
	t.Helper()
	key, err := datakey.Generate()
	assert.NoError(t, err)
	datakey.SetActive(key)
	return key
}

func TestEncryptedFieldRoundTrip(t *testing.T) {
	useDataKey(t)

	data, err := json.Marshal(&encryptedHolder{Secret: NewEncryptedField("my-plain-value")})
	assert.NoError(t, err)
	assert.NotContains(t, string(data), "my-plain-value", "the value must not be stored in the clear")

	restored := &encryptedHolder{}
	assert.NoError(t, json.Unmarshal(data, restored))

	plain, err := restored.Secret.GetPlain()
	assert.NoError(t, err)
	assert.Equal(t, "my-plain-value", plain)
}

// Regression: this used to write an empty string and lose the value without any
// error, because the field holds a plaintext value in `decrypted` while the
// no-key branch marshaled `encrypted`.
func TestEncryptedFieldRefusesToMarshalWithoutKey(t *testing.T) {
	datakey.SetActive(nil)

	_, err := json.Marshal(&encryptedHolder{Secret: NewEncryptedField("my-plain-value")})
	assert.Error(t, err, "marshaling without a data key must fail, not silently drop the value")
}

// An unset field is not an error: plenty of settings leave optional secrets empty.
func TestEncryptedFieldEmptyValueMarshalsWithoutKey(t *testing.T) {
	datakey.SetActive(nil)

	data, err := json.Marshal(&encryptedHolder{})
	assert.NoError(t, err)
	assert.JSONEq(t, `{"secret":""}`, string(data))
}

// A value already encrypted travels untouched, which is what re-saving a setting
// loaded from the database does.
func TestEncryptedFieldKeepsAlreadyEncryptedValue(t *testing.T) {
	useDataKey(t)

	stored, err := json.Marshal(&encryptedHolder{Secret: NewEncryptedField("my-plain-value")})
	assert.NoError(t, err)

	loaded := &encryptedHolder{}
	assert.NoError(t, json.Unmarshal(stored, loaded))

	resaved, err := json.Marshal(loaded)
	assert.NoError(t, err)
	assert.JSONEq(t, string(stored), string(resaved))
}

// A value sealed with a different key must fail loudly rather than come back as
// garbage. AES-GCM is authenticated, so it does.
func TestEncryptedFieldRefusesAnotherKey(t *testing.T) {
	useDataKey(t)
	stored, err := json.Marshal(&encryptedHolder{Secret: NewEncryptedField("my-plain-value")})
	assert.NoError(t, err)

	useDataKey(t) // a different key

	loaded := &encryptedHolder{}
	assert.NoError(t, json.Unmarshal(stored, loaded))
	_, err = loaded.Secret.GetPlain()
	assert.Error(t, err)
}

// Reencrypt is what a data key rotation needs: marshaling alone reuses the
// ciphertext the value was loaded with.
func TestEncryptedFieldReencrypt(t *testing.T) {
	useDataKey(t)
	stored, err := json.Marshal(&encryptedHolder{Secret: NewEncryptedField("my-plain-value")})
	assert.NoError(t, err)

	loaded := &encryptedHolder{}
	assert.NoError(t, json.Unmarshal(stored, loaded))

	resavedAsIs, err := json.Marshal(loaded)
	assert.NoError(t, err)
	assert.JSONEq(t, string(stored), string(resavedAsIs), "a plain re-save keeps the old ciphertext")

	assert.NoError(t, loaded.Secret.Reencrypt())
	rewritten, err := json.Marshal(loaded)
	assert.NoError(t, err)
	assert.NotEqual(t, string(stored), string(rewritten), "the nonce alone makes it differ")

	final := &encryptedHolder{}
	assert.NoError(t, json.Unmarshal(rewritten, final))
	plain, err := final.Secret.GetPlain()
	assert.NoError(t, err)
	assert.Equal(t, "my-plain-value", plain)
}
