package appsettingsdto

import (
	"testing"

	"github.com/stretchr/testify/assert"
	vld "github.com/tiendc/go-validator"
)

func rateLimitErrors(req *HTTPRateLimitConfigReq) []string {
	var out []string
	for _, err := range vld.Validate(req.validate("rateLimitConfig")...) {
		key, _ := err.CustomKey().(string)
		out = append(out, key)
	}
	return out
}

func TestARateLimitTurnedOnMustLimitSomething(t *testing.T) {
	assert.Equal(t, []string{"ERR_VLD_RATE_LIMIT_EMPTY"},
		rateLimitErrors(&HTTPRateLimitConfigReq{Enabled: true}), "on, with nothing set")
	assert.Equal(t, []string{"ERR_VLD_RATE_LIMIT_EMPTY"},
		rateLimitErrors(&HTTPRateLimitConfigReq{Enabled: true, Burst: 50}), "a burst alone limits nothing")

	assert.Empty(t, rateLimitErrors(&HTTPRateLimitConfigReq{Enabled: true, Average: 100}))
	assert.Empty(t, rateLimitErrors(&HTTPRateLimitConfigReq{Enabled: true, MaxInFlightReq: 20}))
	assert.Empty(t, rateLimitErrors(&HTTPRateLimitConfigReq{}), "off, nothing to check")
	assert.Empty(t, rateLimitErrors(nil))
}

func TestARateLimitIsNotNegative(t *testing.T) {
	assert.NotEmpty(t, rateLimitErrors(&HTTPRateLimitConfigReq{Enabled: true, Average: 100, Burst: -1}))
}
