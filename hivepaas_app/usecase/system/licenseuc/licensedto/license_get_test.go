package licensedto

import (
	"encoding/json"
	"maps"
	"slices"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/license"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/licenseservice"
)

const (
	customerEmail = "ops@acme.example"
	thisCluster   = "cluster-this"
)

var (
	issuedAt  = time.Date(2026, 11, 1, 0, 0, 0, 0, time.UTC)
	expiresAt = time.Date(2026, 12, 1, 0, 0, 0, 0, time.UTC)
)

func hiveKey() *license.Payload {
	return &license.Payload{
		Version:   license.PayloadVersion,
		LicenseID: "lic_01",
		Customer:  license.Customer{ID: "acc_1", Name: "Acme", Email: customerEmail},
		Plan:      base.LicensePlanHive,
		Limits:    map[string]int{"nodes": 10, "users": 10},
		Installations: []string{
			thisCluster, "cluster-other",
		},
		IssuedAt:  issuedAt,
		ExpiresAt: expiresAt,
		GraceDays: 14,
	}
}

func someUsage() *licenseservice.Usage {
	return &licenseservice.Usage{Used: map[base.LicenseLimit]int{
		base.LicenseLimitNodes:           3,
		base.LicenseLimitUsers:           4,
		base.LicenseLimitApps:            25,
		base.LicenseLimitFunctions:       2,
		base.LicenseLimitPreviews:        1,
		base.LicenseLimitSignInProviders: 1,
	}}
}

// The wire shape of a key in force, field by field. The customer is its name
// alone, and a kind the plan does not limit has max null.
func TestTransformLicenseOfAKeyInForce(t *testing.T) {
	key := hiveKey()
	state := &license.State{
		Status:         base.LicenseStatusValid,
		Plan:           key.Plan,
		Limits:         license.LimitsOf(key),
		Key:            key,
		InstallationID: thisCluster,
		GraceEndsAt:    key.GraceEndsAt(),
	}

	body, err := json.Marshal(TransformLicense(state, someUsage()))

	assert.NoError(t, err)
	assert.JSONEq(t, `{
		"status": "valid",
		"plan": "hive",
		"keyPlan": "hive",
		"trial": false,
		"development": false,
		"licenseId": "lic_01",
		"customerName": "Acme",
		"installationId": "cluster-this",
		"installations": ["cluster-this", "cluster-other"],
		"issuedAt": "2026-11-01T00:00:00Z",
		"expiresAt": "2026-12-01T00:00:00Z",
		"graceEndsAt": "2026-12-15T00:00:00Z",
		"limits": [
			{"limit": "nodes", "max": 10, "used": 3},
			{"limit": "users", "max": 10, "used": 4},
			{"limit": "apps", "max": null, "used": 25},
			{"limit": "functions", "max": null, "used": 2},
			{"limit": "previews", "max": null, "used": 1},
			{"limit": "sign-in-providers", "max": null, "used": 1}
		]
	}`, string(body))
	assert.NotContains(t, string(body), customerEmail)
	assert.NotContains(t, string(body), "acc_1")
}

// No key is Bee: every limit is there with Bee's max, sign-in providers at 0,
// and none of a key's fields is sent.
func TestTransformLicenseWithNoKey(t *testing.T) {
	state := license.Evaluate(&license.EvalInput{Now: issuedAt})

	resp := TransformLicense(state, someUsage())

	assert.Equal(t, base.LicenseStatusNone, resp.Status)
	assert.Equal(t, base.LicensePlanBee, resp.Plan)
	limits := map[base.LicenseLimit]*int{}
	for i, item := range resp.Limits {
		assert.Equal(t, base.AllLicenseLimits[i], item.Limit, "every limit, in order")
		limits[item.Limit] = item.Max
		assert.Equal(t, someUsage().Used[item.Limit], item.Used)
	}
	assert.Len(t, resp.Limits, len(base.AllLicenseLimits))
	for limit, maxCount := range license.BeeLimits {
		if assert.NotNil(t, limits[limit], limit) {
			assert.Equal(t, maxCount, *limits[limit], limit)
		}
	}
	assert.Equal(t, 0, *limits[base.LicenseLimitSignInProviders])

	body, err := json.Marshal(resp)
	assert.NoError(t, err)
	var fields map[string]any
	assert.NoError(t, json.Unmarshal(body, &fields))
	assert.ElementsMatch(t, []string{"status", "plan", "trial", "development", "installationId", "limits"},
		slices.Collect(maps.Keys(fields)), "no key, no key fields, and no zero times")
}

// A key that is not in force - here for another installation - is still
// described, so the page can say which key it is and why it does not hold, and
// still never with the email.
func TestTransformLicenseOfAKeyNotInForce(t *testing.T) {
	key := hiveKey()
	key.Installations = []string{"cluster-other"}
	key.Trial = true
	state := &license.State{
		Status:         base.LicenseStatusOtherInstallation,
		Plan:           base.LicensePlanBee,
		Limits:         maps.Clone(license.BeeLimits),
		Key:            key,
		Development:    true,
		InstallationID: thisCluster,
	}

	resp := TransformLicense(state, nil)

	assert.Equal(t, base.LicensePlanBee, resp.Plan)
	assert.Equal(t, base.LicensePlanHive, resp.KeyPlan)
	assert.True(t, resp.Trial)
	assert.True(t, resp.Development)
	assert.Equal(t, []string{"cluster-other"}, resp.Installations)
	assert.Nil(t, resp.GraceEndsAt, "a key for another installation has no grace here")
	for _, item := range resp.Limits {
		assert.NotNil(t, item.Max, "Bee limits every kind")
		assert.Zero(t, item.Used, "with no usage, nothing is counted")
	}
	body, err := json.Marshal(resp)
	assert.NoError(t, err)
	assert.NotContains(t, string(body), customerEmail)
}

// An invalid key has no payload, only the reason.
func TestTransformLicenseOfAnInvalidKey(t *testing.T) {
	state := &license.State{
		Status: base.LicenseStatusInvalid,
		Plan:   base.LicensePlanBee,
		Limits: maps.Clone(license.BeeLimits),
		Reason: "neither base64 nor JSON",
	}

	resp := TransformLicense(state, someUsage())

	assert.Equal(t, "neither base64 nor JSON", resp.Reason)
	assert.Empty(t, resp.LicenseID)
	assert.Empty(t, resp.KeyPlan)
	assert.Nil(t, resp.ExpiresAt)
}
