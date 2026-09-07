package basedto

import (
	"testing"

	"github.com/stretchr/testify/assert"
	vld "github.com/tiendc/go-validator"
)

// vld.Validate returns a slice type that satisfies error even when it holds
// nothing, so the assertions below count entries rather than test for nil.
func validateIPs(values []string, minLen int) vld.Errors {
	return vld.Validate(ValidateIPOrCIDRSlice(values, minLen, "trustedIPs")...)
}

func TestValidateIPOrCIDRSlice(t *testing.T) {
	t.Run("accepts addresses and CIDR blocks", func(t *testing.T) {
		assert.Empty(t, validateIPs([]string{
			"10.0.0.0/8", "173.245.48.0/20", "1.1.1.1", "2606:4700::", "2400:cb00::/32",
		}, 1))
	})

	// These reach Traefik as a command-line argument: one it cannot parse keeps
	// the proxy from starting, so a typo here is an outage, not a field error.
	t.Run("rejects what Traefik could not parse", func(t *testing.T) {
		for _, value := range []string{"not-an-address", "10.0.0.0/33", "10.0.0.256", "", " "} {
			assert.NotEmpty(t, validateIPs([]string{value}, 1), "value %q", value)
		}
	})

	t.Run("rejects an empty list when one is required", func(t *testing.T) {
		assert.NotEmpty(t, validateIPs(nil, 1))
	})

	t.Run("rejects duplicates", func(t *testing.T) {
		assert.NotEmpty(t, validateIPs([]string{"10.0.0.0/8", "10.0.0.0/8"}, 1))
	})

	t.Run("an empty list is fine when none is required", func(t *testing.T) {
		assert.Empty(t, validateIPs(nil, 0))
	})
}
