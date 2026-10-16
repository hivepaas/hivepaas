package license

import (
	"encoding/base64"
	"strings"
	"unicode"
)

// MaxKeyTextLen bounds the text taken as a key. A key is a few kilobytes - the
// ML-DSA-65 signature alone is 3309 bytes - so this leaves room for an Apiary
// key naming many installations, and keeps anything much larger from being
// decoded at all.
const MaxKeyTextLen = 32 << 10

// NormalizeKeyText removes every whitespace character from s. Nothing in an
// envelope holds any - its strings are base64, hex, key ids and algorithm names -
// so this changes nothing a key says, and a key wrapped by an email client or
// indented in a file can be pasted as it is.
func NormalizeKeyText(s string) string {
	return strings.Map(func(r rune) rune {
		if unicode.IsSpace(r) {
			return -1
		}
		return r
	}, s)
}

// DecodeKeyText returns the envelope a key's text holds, once its whitespace is
// removed: the envelope's JSON as base64, which is the form keys are issued in,
// or the envelope's JSON as it is. Padding may be left off the base64.
func DecodeKeyText(s string) ([]byte, error) {
	s = NormalizeKeyText(s)
	if s == "" {
		return nil, invalid("no key")
	}
	if len(s) > MaxKeyTextLen {
		return nil, invalid("longer than %d characters", MaxKeyTextLen)
	}
	if strings.HasPrefix(s, "{") {
		return []byte(s), nil
	}
	envelope, err := base64.RawStdEncoding.DecodeString(strings.TrimRight(s, "="))
	if err != nil {
		return nil, invalid("neither base64 nor JSON")
	}
	return envelope, nil
}
