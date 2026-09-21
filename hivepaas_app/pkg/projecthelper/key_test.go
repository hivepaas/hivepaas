package projecthelper

import (
	"regexp"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

// dnsLabelPattern is what a host name label may be: letters, digits and hyphens,
// not starting or ending with a hyphen, at most 63 long.
var dnsLabelPattern = regexp.MustCompile(`^[a-z0-9]([a-z0-9-]{0,61}[a-z0-9])?$`)

func TestCalcAppKeyKeepsHyphens(t *testing.T) {
	assert.Equal(t, "blog-db", CalcAppKey("blog-db"))
	assert.Equal(t, "my-shop", CalcAppKey("My Shop"))
	assert.Equal(t, "t-apprise", CalcAppKey("t_apprise"))
}

func TestCalcAppKeyIsAHostName(t *testing.T) {
	for _, name := range []string{
		"blog-db",
		"My Shop!",
		"  spaced   out  ",
		"Übergröße café",
		"t_apprise",
		strings.Repeat("a", 200),
		// 62 letters and a word: the cut at 63 falls right after the separator.
		strings.Repeat("a", 62) + " tail",
	} {
		key := CalcAppKey(name)
		assert.Regexp(t, dnsLabelPattern, key, "name %q gave key %q", name, key)
	}
}

func TestCalcAppKeyIsCutToALabel(t *testing.T) {
	key := CalcAppKey(strings.Repeat("a", 62) + " tail")
	assert.Len(t, key, 62)
	assert.False(t, strings.HasSuffix(key, "-"))

	assert.Len(t, CalcAppKey(strings.Repeat("b", 100)), AppKeyMaxLen)
}
