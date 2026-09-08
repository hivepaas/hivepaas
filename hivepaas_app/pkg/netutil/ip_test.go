package netutil

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestIPAllowed(t *testing.T) {
	t.Run("an unparseable caller matches nothing", func(t *testing.T) {
		assert.False(t, IPAllowed([]string{"0.0.0.0/0"}, "not-an-address"))
	})

	t.Run("an unparseable entry is skipped, not fatal", func(t *testing.T) {
		assert.True(t, IPAllowed([]string{"garbage", "203.0.113.7"}, "203.0.113.7"))
	})

	t.Run("a v4 address is not inside a v6 block", func(t *testing.T) {
		assert.False(t, IPAllowed([]string{"2001:db8::/32"}, "203.0.113.7"))
	})
}
