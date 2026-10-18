package license

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/hivepaas/hivepaas/hivepaas_app/base"
)

// A key grants the features it lists while it is in force, and only the ones this
// binary knows; Bee grants none.
func TestEvaluateFeatures(t *testing.T) {
	prod := newPair(t, "lic-2026")
	trust := &Trust{Prod: publicKeysOf(t, prod...)}
	key := func(change func(p *Payload)) string {
		p := validPayload()
		p.Features = []string{"managed-databases", "a-feature-from-a-newer-portal"}
		if change != nil {
			change(p)
		}
		return token(seal(t, marshal(t, p), Context, prod...))
	}
	expiresAt := validPayload().ExpiresAt

	for _, tc := range []struct {
		name    string
		keyText string
		now     time.Time
		granted []base.LicenseFeature
	}{
		{name: "no key"},
		{name: "a key in force", keyText: key(nil), granted: []base.LicenseFeature{base.LicenseFeatureManagedDatabases}},
		{name: "a key in grace", keyText: key(nil), now: expiresAt.Add(day),
			granted: []base.LicenseFeature{base.LicenseFeatureManagedDatabases}},
		{name: "a key past grace", keyText: key(nil), now: expiresAt.Add(30 * day)},
		{name: "a key for another installation", keyText: key(func(p *Payload) { p.Installations = []string{"other"} })},
		{name: "a key listing none", keyText: key(func(p *Payload) { p.Features = nil })},
	} {
		t.Run(tc.name, func(t *testing.T) {
			in := &EvalInput{KeyText: tc.keyText, InstallationID: testCluster, Now: tc.now}
			if in.Now.IsZero() {
				in.Now = testNow
			}
			if in.KeyText != "" {
				in.Verified, in.VerifyErr = Verify(trust, in.KeyText)
				assert.NoError(t, in.VerifyErr)
			}

			state := Evaluate(in)

			assert.Equal(t, tc.granted, state.Features)
			assert.Equal(t, len(tc.granted) > 0, state.Grants(string(base.LicenseFeatureManagedDatabases)))
			assert.False(t, state.Grants("a-feature-from-a-newer-portal"), "unknown here, so not granted")
		})
	}
}

func TestEveryFeatureHasAName(t *testing.T) {
	for _, feature := range base.AllLicenseFeatures {
		assert.NotEmpty(t, FeatureNameMsgID(feature), feature)
	}
	assert.Empty(t, FeatureNameMsgID("unknown"))
}
