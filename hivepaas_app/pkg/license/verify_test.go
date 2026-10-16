package license

import (
	"bytes"
	"crypto/ed25519"
	"crypto/mldsa"
	"crypto/rand"
	"crypto/sha256"
	"crypto/x509"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"encoding/pem"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/releasesig"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/translation"
)

const (
	testEmail        = "ops@acme.example"
	testCustomerName = "Acme"
	testCluster      = "cluster-a"
)

var testNow = time.Date(2026, 11, 10, 12, 0, 0, 0, time.UTC)

// testKey is one half of a signing pair, the way the portal holds it.
type testKey struct {
	id   string
	alg  string
	pem  []byte
	sign func(data []byte, context string) []byte
}

// newPair makes an ed25519 key and an ML-DSA-65 key, named <prefix>-ed and
// <prefix>-ml.
func newPair(t *testing.T, prefix string) []*testKey {
	t.Helper()
	edPub, edPriv, err := ed25519.GenerateKey(rand.Reader)
	assert.NoError(t, err)
	mlPriv, err := mldsa.GenerateKey(mldsa.MLDSA65())
	assert.NoError(t, err)
	return []*testKey{
		{id: prefix + "-ed", alg: releasesig.AlgEd25519, pem: pemOf(t, edPub),
			sign: func(data []byte, context string) []byte {
				sig, err := edPriv.Sign(nil, data, &ed25519.Options{Context: context})
				assert.NoError(t, err)
				return sig
			}},
		{id: prefix + "-ml", alg: releasesig.AlgMLDSA65, pem: pemOf(t, mlPriv.PublicKey()),
			sign: func(data []byte, context string) []byte {
				sig, err := mlPriv.SignDeterministic(data, &mldsa.Options{Context: context})
				assert.NoError(t, err)
				return sig
			}},
	}
}

func pemOf(t *testing.T, pub any) []byte {
	t.Helper()
	der, err := x509.MarshalPKIXPublicKey(pub)
	assert.NoError(t, err)
	return pem.EncodeToMemory(&pem.Block{Type: "PUBLIC KEY", Bytes: der})
}

func publicKeysOf(t *testing.T, keys ...*testKey) releasesig.PublicKeys {
	t.Helper()
	pems := map[string][]byte{}
	for _, key := range keys {
		pems[key.id] = key.pem
	}
	parsed, err := Purpose.ParsePublicKeys(pems)
	assert.NoError(t, err)
	return parsed
}

// seal signs data under context with keys, and returns the envelope's JSON.
func seal(t *testing.T, data []byte, context string, keys ...*testKey) []byte {
	t.Helper()
	sum := sha256.Sum256(data)
	env := releasesig.Envelope{
		Payload: base64.StdEncoding.EncodeToString(data),
		SHA256:  hex.EncodeToString(sum[:]),
	}
	for _, key := range keys {
		env.Signatures = append(env.Signatures, releasesig.Entry{
			KeyID: key.id, Algorithm: key.alg, Sig: base64.StdEncoding.EncodeToString(key.sign(data, context)),
		})
	}
	content, err := json.Marshal(env)
	assert.NoError(t, err)
	return content
}

// token is a key's text the way keys are issued: the envelope's JSON, base64.
func token(envelope []byte) string {
	return base64.StdEncoding.EncodeToString(envelope)
}

func validPayload() *Payload {
	return &Payload{
		Version:        PayloadVersion,
		LicenseID:      "lic_01",
		SubscriptionID: "sub_01",
		Customer:       Customer{ID: "acc_01", Name: testCustomerName, Email: testEmail},
		Plan:           base.LicensePlanHive,
		Limits:         map[string]int{"nodes": 10, "users": 10},
		Installations:  []string{testCluster},
		IssuedAt:       testNow.Add(-day),
		ExpiresAt:      testNow.Add(29 * day),
		GraceDays:      14,
	}
}

func marshal(t *testing.T, v any) []byte {
	t.Helper()
	data, err := json.Marshal(v)
	assert.NoError(t, err)
	return data
}

// assertRefused checks a refusal is ErrLicenseInvalid, and that what a person
// would read of it holds nothing of the customer's.
func assertRefused(t *testing.T, err error) {
	t.Helper()
	if !assert.ErrorIs(t, err, hperrors.ErrLicenseInvalid) {
		return
	}
	detail := hperrors.GetErrorDetail(err, translation.LangEn)
	assert.NotContains(t, detail, testEmail)
	assert.NotContains(t, detail, testCustomerName)
	assert.NotContains(t, err.Error(), testEmail)
}

func TestVerifyAcceptsEachForm(t *testing.T) {
	prod := newPair(t, "lic-2026")
	trust := &Trust{Prod: publicKeysOf(t, prod...)}
	data := marshal(t, validPayload())
	envelope := seal(t, data, Context, prod...)

	wrapped := func(s string, width int) string {
		var b strings.Builder
		for len(s) > width {
			b.WriteString(s[:width] + "\r\n")
			s = s[width:]
		}
		return "  \t" + b.String() + s + "\n\n"
	}
	var indented bytes.Buffer
	assert.NoError(t, json.Indent(&indented, envelope, "", "  "))

	for name, text := range map[string]string{
		"token":                         token(envelope),
		"token without padding":         strings.TrimRight(token(envelope), "="),
		"token wrapped by email":        wrapped(token(envelope), 76),
		"token with no-break spaces":    " " + token(envelope) + " ",
		"envelope JSON":                 string(envelope),
		"envelope JSON, one per line":   indented.String(),
		"envelope JSON wrapped by mail": wrapped(string(envelope), 60),
	} {
		t.Run(name, func(t *testing.T) {
			verified, err := Verify(trust, text)
			if !assert.NoError(t, err) {
				return
			}
			assert.False(t, verified.Development)
			assert.Equal(t, validPayload(), verified.Payload)
		})
	}
}

func TestVerifyRefuses(t *testing.T) {
	prod := newPair(t, "lic-2026")
	trust := &Trust{Prod: publicKeysOf(t, prod...), Dev: publicKeysOf(t, newPair(t, "dev")...)}
	data := marshal(t, validPayload())
	other := marshal(t, func() *Payload { p := validPayload(); p.Plan = base.LicensePlanApiary; return p }())
	ed, ml := prod[0], prod[1]
	entry := func(key *testKey, signed []byte, context string) releasesig.Entry {
		return releasesig.Entry{KeyID: key.id, Algorithm: key.alg,
			Sig: base64.StdEncoding.EncodeToString(key.sign(signed, context))}
	}
	envelopeOf := func(payload, sumOf []byte, entries ...releasesig.Entry) string {
		sum := sha256.Sum256(sumOf)
		return token(marshal(t, releasesig.Envelope{Payload: base64.StdEncoding.EncodeToString(payload),
			SHA256: hex.EncodeToString(sum[:]), Signatures: entries}))
	}
	tamperedSig := entry(ml, data, Context)
	tamperedSig.Sig = base64.StdEncoding.EncodeToString(append(ml.sign(data, Context)[1:], 0))

	for name, text := range map[string]string{
		"payload replaced, signatures kept": envelopeOf(other, other,
			entry(ed, data, Context), entry(ml, data, Context)),
		"payload replaced, sha256 kept": envelopeOf(other, data, entry(ed, data, Context), entry(ml, data, Context)),
		"sha256 that does not match":    envelopeOf(data, other, entry(ed, data, Context), entry(ml, data, Context)),
		"a signature altered":           envelopeOf(data, data, entry(ed, data, Context), tamperedSig),
		"signed as a release":           token(seal(t, data, releasesig.Context, prod...)),
		"signed without context":        token(seal(t, data, "", prod...)),
		"signed by keys nobody trusts":  token(seal(t, data, Context, newPair(t, "lic-2027")...)),
		"other keys under trusted ids":  token(seal(t, data, Context, newPair(t, "lic-2026")...)),
		"no ed25519 signature":          token(seal(t, data, Context, ml)),
		"no ML-DSA signature":           token(seal(t, data, Context, ed)),
		"no signatures":                 token(seal(t, data, Context)),
		"empty":                         "",
		"whitespace only":               " \n\t ",
		"neither base64 nor JSON":       "this is not a license key",
		"base64 of something else":      base64.StdEncoding.EncodeToString([]byte("hello")),
		"malformed JSON":                `{"payload": `,
		"too long":                      strings.Repeat("A", MaxKeyTextLen+4),
	} {
		t.Run(name, func(t *testing.T) {
			verified, err := Verify(trust, text)
			assert.Nil(t, verified)
			assertRefused(t, err)
		})
	}
}

// Anybody can sign with the development keys: what they signed counts only
// where the caller says development keys do, and is refused elsewhere for what
// it is.
func TestVerifyDevelopmentKeys(t *testing.T) {
	prod, dev := newPair(t, "lic-2026"), newPair(t, "dev")
	prodKeys, devKeys := publicKeysOf(t, prod...), publicKeysOf(t, dev...)
	data := marshal(t, validPayload())
	byDev := token(seal(t, data, Context, dev...))

	verified, err := Verify(&Trust{Prod: prodKeys, Dev: devKeys, AllowDev: true}, byDev)
	if !assert.NoError(t, err) {
		return
	}
	assert.True(t, verified.Development)

	_, err = Verify(&Trust{Prod: prodKeys, Dev: devKeys}, byDev)
	assertRefused(t, err)
	assert.Contains(t, hperrors.GetErrorDetail(err, translation.LangEn),
		"signed by a development key, accepted only in development")

	// No production keys at all, as in a binary built before the first pair.
	verified, err = Verify(&Trust{Dev: devKeys, AllowDev: true}, byDev)
	if !assert.NoError(t, err) {
		return
	}
	assert.True(t, verified.Development)

	verified, err = Verify(&Trust{Prod: prodKeys, Dev: devKeys, AllowDev: true}, token(seal(t, data, Context, prod...)))
	if !assert.NoError(t, err) {
		return
	}
	assert.False(t, verified.Development, "a production key is one whether or not development keys count")

	// A key that names the development keys is judged against them: the
	// refusal says what is wrong with it, not that no production key signed it.
	tampered := seal(t, data, Context, dev...)
	tampered = []byte(strings.Replace(string(tampered), `"sha256":"`, `"sha256":"0`, 1))
	_, err = Verify(&Trust{Prod: prodKeys, Dev: devKeys, AllowDev: true}, string(tampered))
	assertRefused(t, err)
	assert.Contains(t, hperrors.GetErrorDetail(err, translation.LangEn), "sha256 does not match")

	_, err = Verify(&Trust{Dev: devKeys}, byDev)
	assertRefused(t, err)
}

// What a key may say, beyond its signatures.
func TestVerifyPayloadRules(t *testing.T) {
	prod := newPair(t, "lic-2026")
	trust := &Trust{Prod: publicKeysOf(t, prod...)}
	signed := func(data []byte) string { return token(seal(t, data, Context, prod...)) }
	with := func(change func(p *Payload)) []byte {
		p := validPayload()
		change(p)
		return marshal(t, p)
	}

	refused := map[string][]byte{
		"version 2":                 with(func(p *Payload) { p.Version = 2 }),
		"version missing":           with(func(p *Payload) { p.Version = 0 }),
		"no license id":             with(func(p *Payload) { p.LicenseID = "" }),
		"plan bee":                  with(func(p *Payload) { p.Plan = base.LicensePlanBee }),
		"unknown plan":              with(func(p *Payload) { p.Plan = "swarm" }),
		"no plan":                   with(func(p *Payload) { p.Plan = "" }),
		"no installations":          with(func(p *Payload) { p.Installations = nil }),
		"an empty installation":     with(func(p *Payload) { p.Installations = []string{testCluster, ""} }),
		"no issuedAt":               with(func(p *Payload) { p.IssuedAt = time.Time{} }),
		"no expiresAt":              with(func(p *Payload) { p.ExpiresAt = time.Time{} }),
		"expires when issued":       with(func(p *Payload) { p.ExpiresAt = p.IssuedAt }),
		"expires before issued":     with(func(p *Payload) { p.ExpiresAt = p.IssuedAt.Add(-time.Second) }),
		"negative grace":            with(func(p *Payload) { p.GraceDays = -1 }),
		"grace over a year":         with(func(p *Payload) { p.GraceDays = 366 }),
		"a negative limit":          with(func(p *Payload) { p.Limits["users"] = -1 }),
		"not an object":             []byte(`[1,2]`),
		"a field of the wrong type": []byte(`{"version":"1","customer":{"email":"` + testEmail + `"}}`),
		"a date that is no date":    []byte(`{"version":1,"issuedAt":"` + testEmail + `"}`),
	}
	for name, data := range refused {
		t.Run("refuses "+name, func(t *testing.T) {
			_, err := Verify(trust, signed(data))
			assertRefused(t, err)
		})
	}

	accepted := map[string][]byte{
		"no grace":   with(func(p *Payload) { p.GraceDays = 0 }),
		"year grace": with(func(p *Payload) { p.GraceDays = maxGraceDays }),
		"no limits":  with(func(p *Payload) { p.Limits = nil }),
		"a limit of zero": with(func(p *Payload) {
			p.Limits["previews"] = 0
		}),
		// A key from a newer portal: what this binary does not know is passed over.
		"an unknown limit": with(func(p *Payload) { p.Limits["gpus"] = 4 }),
		"an unknown field": []byte(strings.Replace(string(marshal(t, validPayload())), `{`, `{"keyId":"lic-2026",`, 1)),
		"several installations": with(func(p *Payload) {
			p.Plan = base.LicensePlanApiary
			p.Installations = []string{"cluster-x", testCluster}
		}),
	}
	for name, data := range accepted {
		t.Run("accepts "+name, func(t *testing.T) {
			_, err := Verify(trust, signed(data))
			assert.NoError(t, err)
		})
	}
}
