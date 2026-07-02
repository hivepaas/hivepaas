package translation

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestLocalize(t *testing.T) {
	// 1. Simple localization (English)
	msg, err := Localize(LangEn, "ERR_UNAUTHORIZED")
	assert.NoError(t, err)
	assert.Equal(t, "Unauthorized", msg)

	// 2. Localization with template data
	msg, err = LocalizeEx(LangEn, "ERR_ARGUMENT_INVALID", map[string]any{"Name": "Email"})
	assert.NoError(t, err)
	assert.Equal(t, "Email is invalid", msg)

	// 3. Fallback to default language (English) for unsupported language
	msg, err = Localize(Lang("fr"), "ERR_BAD_REQUEST")
	assert.NoError(t, err)
	assert.Equal(t, "Bad request", msg)

	// 4. Missing ID should return error from go-i18n
	_, err = Localize(LangEn, "NON_EXISTENT_ID")
	assert.Error(t, err)
}

func TestLocalizeOrKey(t *testing.T) {
	// 1. Existing ID
	assert.Equal(t, "Unauthorized", LocalizeOrKey(LangEn, "ERR_UNAUTHORIZED", nil))

	// 2. Localization with template data
	assert.Equal(t, "Panic: something went wrong",
		LocalizeOrKey(LangEn, "ERR_PANIC", map[string]any{"Error": "something went wrong"}))

	// 3. Missing ID should return the key itself
	assert.Equal(t, "NON_EXISTENT_ID", LocalizeOrKey(LangEn, "NON_EXISTENT_ID", nil))
}

func TestIsAvailable(t *testing.T) {
	assert.True(t, IsAvailable(LangEn))
	assert.False(t, IsAvailable(Lang("vi")))
}

func TestLang_String(t *testing.T) {
	assert.Equal(t, "en", LangEn.String())
}

func TestGetDefaultLang(t *testing.T) {
	assert.Equal(t, LangEn, GetDefaultLang())
}
