package license

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNormalizeKeyText(t *testing.T) {
	assert.Equal(t, "eyJwYXls+/=", NormalizeKeyText(" eyJw\r\nYXls\t+/=\n "))
	assert.Equal(t, `{"payload":"x"}`, NormalizeKeyText("{\n  \"payload\": \"x\"\n}\n"))
	assert.Empty(t, NormalizeKeyText(" \n\t"))
}

func TestDecodeKeyText(t *testing.T) {
	envelope := `{"payload":"eA==","sha256":"0","signatures":[]}`
	padded := token([]byte(envelope))
	assert.True(t, strings.HasSuffix(padded, "="), "the case below needs a token that has padding")

	for name, text := range map[string]string{
		"a token":                 padded,
		"a token without padding": strings.TrimRight(padded, "="),
		"a wrapped token":         padded[:10] + "\n" + padded[10:],
		"the envelope's JSON":     envelope,
		"the JSON indented":       "{\n  \"payload\": \"eA==\",\n  \"sha256\": \"0\",\n  \"signatures\": []\n}\n",
	} {
		t.Run(name, func(t *testing.T) {
			decoded, err := DecodeKeyText(text)
			assert.NoError(t, err)
			assert.Equal(t, envelope, string(decoded))
		})
	}

	for name, text := range map[string]string{
		"nothing":         "",
		"only whitespace": "\n \t",
		"not base64":      "not a key!",
		"too long":        strings.Repeat("QUFB", MaxKeyTextLen/4+1),
	} {
		t.Run("refuses "+name, func(t *testing.T) {
			_, err := DecodeKeyText(text)
			assertRefused(t, err)
		})
	}

	_, err := DecodeKeyText(strings.Repeat("QUFB", MaxKeyTextLen/4) + strings.Repeat("\n", 100))
	assert.NoError(t, err, "whitespace does not count against the length")
}
