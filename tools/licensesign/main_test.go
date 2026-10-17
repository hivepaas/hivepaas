package main

import (
	"crypto"
	"crypto/ed25519"
	"crypto/mldsa"
	"crypto/rand"
	"crypto/x509"
	"encoding/base64"
	"encoding/json"
	"encoding/pem"
	"errors"
	"maps"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"
	"time"

	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/license"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/license/licensekeys"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/releasesig"
)

var testNow = time.Date(2026, 11, 1, 9, 30, 0, 0, time.UTC)

func newKeys(t *testing.T, prefix string) []*privateKey {
	t.Helper()
	_, ed, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		t.Fatal(err)
	}
	ml, err := mldsa.GenerateKey(mldsa.MLDSA65())
	if err != nil {
		t.Fatal(err)
	}
	return []*privateKey{
		{id: prefix + "-ed", alg: algEd25519, key: ed},
		{id: prefix + "-ml", alg: algMLDSA65, key: ml},
	}
}

func publicOf(keys []*privateKey) []*publicKey {
	pubs := make([]*publicKey, 0, len(keys))
	for _, key := range keys {
		pubs = append(pubs, &publicKey{id: key.id, alg: key.alg, key: key.key.Public()})
	}
	return pubs
}

func pemPublicKey(t *testing.T, pub crypto.PublicKey) []byte {
	t.Helper()
	der, err := x509.MarshalPKIXPublicKey(pub)
	if err != nil {
		t.Fatal(err)
	}
	return pem.EncodeToMemory(&pem.Block{Type: "PUBLIC KEY", Bytes: der})
}

// appTrust is the app's trust in keys, as its production keys.
func appTrust(t *testing.T, keys []*privateKey) *license.Trust {
	t.Helper()
	pems := map[string][]byte{}
	for _, key := range keys {
		pems[key.id] = pemPublicKey(t, key.key.Public())
	}
	parsed, err := license.Purpose.ParsePublicKeys(pems)
	if err != nil {
		t.Fatal(err)
	}
	return &license.Trust{Prod: parsed}
}

// devKeys are the committed development keys.
func devKeys(t *testing.T) []*privateKey {
	t.Helper()
	keys, err := readPrivateKeys([]string{filepath.Join("devkeys", "dev-ed.key"), filepath.Join("devkeys", "dev-ml.key")})
	if err != nil {
		t.Fatal(err)
	}
	return keys
}

func colonyOptions() *issueOptions {
	return &issueOptions{
		plan: "colony", installations: []string{"cluster-e2e"}, days: 30, graceDays: defaultGraceDays, nodes: -1, users: -1,
		licenseID: "lic_e2e", customer: Customer{Name: "HivePaaS e2e", Email: "admin@e2e.localhost"},
	}
}

// issueToFile runs what issue does after reading its flags, into a file.
func issueToFile(t *testing.T, keys []*privateKey, opts *issueOptions, format string) string {
	t.Helper()
	payload, err := issuePayload(opts, testNow)
	if err != nil {
		t.Fatal(err)
	}
	data, err := json.Marshal(payload)
	if err != nil {
		t.Fatal(err)
	}
	out := filepath.Join(t.TempDir(), "key.txt")
	if err = signAndWrite(keys, data, payload, out, format); err != nil {
		t.Fatal(err)
	}
	content, err := os.ReadFile(out)
	if err != nil {
		t.Fatal(err)
	}
	return string(content)
}

func TestSignOpen(t *testing.T) {
	keys := newKeys(t, "lic-2026")
	data := []byte(`{"version":1}`)

	env, err := Sign(keys, data)
	if err != nil {
		t.Fatal(err)
	}
	for _, format := range []string{formatToken, formatJSON} {
		opened, err := Open(publicOf(keys), FormatKey(env, format))
		if err != nil {
			t.Fatalf("a fresh %s must open: %v", format, err)
		}
		if string(opened) != string(data) {
			t.Fatalf("the %s must carry the payload byte for byte", format)
		}
	}
	token := FormatKey(env, formatToken)
	if strings.Count(string(token), "\n") != 1 || !strings.HasSuffix(string(token), "\n") {
		t.Fatal("a token is one line")
	}

	changed := *env
	changed.Payload = base64.StdEncoding.EncodeToString([]byte(`{"version":2}`))
	if _, err = Open(publicOf(keys), FormatKey(&changed, formatToken)); err == nil {
		t.Fatal("a changed payload must not open")
	}
	if _, err = Open(publicOf(newKeys(t, "lic-2026")), token); err == nil {
		t.Fatal("other keys under the same ids must not open it")
	}
	if _, err = Open(publicOf(keys[:1]), token); err == nil {
		t.Fatal("trusting only one algorithm's key must not satisfy both")
	}
	if _, err = Sign(append(keys, newKeys(t, "x")[0]), data); err == nil {
		t.Fatal("two keys of one algorithm must be refused")
	}
}

// Both algorithms are required: the tool will not sign with one, and neither
// it nor the app opens a key that carries one signature.
func TestRefusesAnEnvelopeMissingAnAlgorithm(t *testing.T) {
	keys := newKeys(t, "lic-2026")
	data, err := json.Marshal(mustIssue(t, colonyOptions()))
	if err != nil {
		t.Fatal(err)
	}
	for _, one := range keys {
		if _, err = Sign([]*privateKey{one}, data); err == nil {
			t.Fatalf("signing with %s alone must be refused", one.alg)
		}
	}

	env, err := Sign(keys, data)
	if err != nil {
		t.Fatal(err)
	}
	for _, dropped := range requiredAlgorithms {
		partial := *env
		partial.Signatures = slices.DeleteFunc(slices.Clone(env.Signatures), func(entry SigEntry) bool {
			return entry.Algorithm == dropped
		})
		token := FormatKey(&partial, formatToken)
		if _, err = Open(publicOf(keys), token); err == nil {
			t.Fatalf("a key without its %s signature must not open", dropped)
		}
		if _, err = license.Verify(appTrust(t, keys), string(token)); !errors.Is(err, hperrors.ErrLicenseInvalid) {
			t.Fatalf("the app must refuse a key without its %s signature, got %v", dropped, err)
		}
	}
}

// What issue writes, the app takes: signed with the committed development keys,
// a Colony key for the e2e environment's swarm puts Colony's limits in force
// there, and only where development keys count.
func TestIssuedKeyIsInForceInTheApp(t *testing.T) {
	for _, format := range []string{formatToken, formatJSON} {
		text := issueToFile(t, devKeys(t), colonyOptions(), format)

		trust, err := licensekeys.Trusted(true)
		if err != nil {
			t.Fatal(err)
		}
		verified, err := license.Verify(trust, text)
		if err != nil {
			t.Fatalf("the app must take the %s issue writes: %v", format, err)
		}
		state := license.Evaluate(&license.EvalInput{KeyText: text, Verified: verified,
			InstallationID: "cluster-e2e", Now: testNow.Add(time.Hour)})
		if state.Status != base.LicenseStatusValid || state.Plan != base.LicensePlanColony || !state.Development {
			t.Fatalf("got status %s, plan %s, development %v; want valid, colony, true",
				state.Status, state.Plan, state.Development)
		}
		if want := (license.Limits{base.LicenseLimitNodes: 25, base.LicenseLimitUsers: 30}); !maps.Equal(state.Limits,
			want) {
			t.Fatalf("limits %v, want %v", state.Limits, want)
		}
		if !state.GraceEndsAt.Equal(testNow.Add(60 * day)) {
			t.Fatalf("grace ends %s, want 30 days and 30 grace days on", state.GraceEndsAt)
		}

		trust, err = licensekeys.Trusted(false)
		if err != nil {
			t.Fatal(err)
		}
		if _, err = license.Verify(trust, text); !errors.Is(err, hperrors.ErrLicenseInvalid) {
			t.Fatalf("outside development the app must refuse a development key, got %v", err)
		}
	}
}

// The e2e environment signs with the committed private keys and the app
// verifies with the embedded public keys: the two must be one pair.
func TestDevKeysMatchTheEmbeddedPublicKeys(t *testing.T) {
	embedded, err := licensekeys.Dev()
	if err != nil {
		t.Fatal(err)
	}
	keys := devKeys(t)
	committed, err := filepath.Glob(filepath.Join("devkeys", "*.key"))
	if err != nil {
		t.Fatal(err)
	}
	if len(committed) != len(keys) || len(embedded) != len(keys) {
		t.Fatalf("%d private keys committed, %d public keys embedded; want %d of each",
			len(committed), len(embedded), len(keys))
	}
	for _, key := range keys {
		derived := pemPublicKey(t, key.key.Public())
		if string(derived) != string(embedded[key.id]) {
			t.Fatalf("devkeys/%s.key does not derive licensekeys/dev/%s.pub.pem", key.id, key.id)
		}
	}
}

// The tool knows the license context alone: what it signs is not a release.json,
// even when handed the keys the app trusts for releases.
func TestCannotSignARelease(t *testing.T) {
	keys := newKeys(t, "2026")
	env, err := Sign(keys, []byte(`{"stable":{"appVersion":"v9.9.9"}}`))
	if err != nil {
		t.Fatal(err)
	}
	pems := map[string][]byte{}
	for _, key := range keys {
		pems[key.id] = pemPublicKey(t, key.key.Public())
	}
	trusted, err := releasesig.ParsePublicKeys(pems)
	if err != nil {
		t.Fatal(err)
	}
	envelope, err := decodeKeyText(string(FormatKey(env, formatToken)))
	if err != nil {
		t.Fatal(err)
	}
	if _, err = releasesig.Open(trusted, envelope); err == nil {
		t.Fatal("the app must refuse as a release what the tool signs")
	}
}

func TestIssuePayload(t *testing.T) {
	limitsOf := func(opts *issueOptions) map[string]int {
		t.Helper()
		return mustIssue(t, opts).Limits
	}
	for plan, want := range map[string]map[string]int{
		"comb":   {"nodes": 5, "users": 5},
		"hive":   {"nodes": 10, "users": 10},
		"colony": {"nodes": 25, "users": 30},
	} {
		opts := colonyOptions()
		opts.plan = plan
		if got := limitsOf(opts); !maps.Equal(got, want) {
			t.Fatalf("%s limits %v, want %v", plan, got, want)
		}
	}

	opts := colonyOptions()
	opts.nodes = 40
	if got := limitsOf(opts); got["nodes"] != 40 || got["users"] != 30 {
		t.Fatalf("-nodes must override the plan's nodes alone, got %v", got)
	}

	opts = colonyOptions()
	opts.plan = "apiary"
	opts.nodes = 100
	if _, err := issuePayload(opts, testNow); err == nil || !strings.Contains(err.Error(), "-users") {
		t.Fatalf("an apiary key without -users must be refused, got %v", err)
	}
	opts.users = 250
	opts.installations = []string{"cluster-a", "cluster-b"}
	if got := limitsOf(opts); !maps.Equal(got, map[string]int{"nodes": 100, "users": 250}) {
		t.Fatalf("apiary limits %v", got)
	}

	opts = colonyOptions()
	opts.plan = "hive"
	opts.days = 14
	opts.trial = true
	trial := mustIssue(t, opts)
	if trial.GraceDays != 0 || !trial.Trial || !trial.ExpiresAt.Equal(testNow.Add(14*day)) {
		t.Fatalf("a trial has no grace and runs out on its date, got %+v", trial)
	}

	opts = colonyOptions()
	opts.licenseID = ""
	if id := mustIssue(t, opts).LicenseID; !strings.HasPrefix(id, "lic_") || len(id) != len("lic_")+16 {
		t.Fatalf("a license id is made up when none is given, got %q", id)
	}

	for name, change := range map[string]func(*issueOptions){
		"bee":               func(o *issueOptions) { o.plan = "bee" },
		"an unknown plan":   func(o *issueOptions) { o.plan = "swarm" },
		"no installation":   func(o *issueOptions) { o.installations = nil },
		"an empty one":      func(o *issueOptions) { o.installations = []string{""} },
		"no days":           func(o *issueOptions) { o.days = 0 },
		"a negative grace":  func(o *issueOptions) { o.graceDays = -1 },
		"grace over a year": func(o *issueOptions) { o.graceDays = 366 },
	} {
		opts := colonyOptions()
		change(opts)
		if _, err := issuePayload(opts, testNow); err == nil {
			t.Fatalf("issuing with %s must be refused", name)
		}
	}
}

// The payload's rules are kept in two places: what the tool refuses, the app
// refuses too. The tool is stricter in one way, on purpose: it refuses a limit
// name the app would pass over.
func TestValidationMatchesTheApp(t *testing.T) {
	keys := newKeys(t, "lic-2026")
	trust := appTrust(t, keys)
	with := func(change func(p *Payload)) *Payload {
		p := mustIssue(t, colonyOptions())
		change(p)
		return p
	}
	check := func(name string, p *Payload, toolTakes, appTakes bool) {
		t.Helper()
		if err := validatePayload(p); (err == nil) != toolTakes {
			t.Fatalf("%s: the tool takes it: %v, want %v (%v)", name, err == nil, toolTakes, err)
		}
		data, err := json.Marshal(p)
		if err != nil {
			t.Fatal(err)
		}
		env, err := Sign(keys, data)
		if err != nil {
			t.Fatal(err)
		}
		if _, err = license.Verify(trust, string(FormatKey(env, formatToken))); (err == nil) != appTakes {
			t.Fatalf("%s: the app takes it: %v, want %v (%v)", name, err == nil, appTakes, err)
		}
	}

	check("a colony key", with(func(*Payload) {}), true, true)
	check("no grace", with(func(p *Payload) { p.GraceDays = 0 }), true, true)
	for name, change := range map[string]func(*Payload){
		"version 2":           func(p *Payload) { p.Version = 2 },
		"no license id":       func(p *Payload) { p.LicenseID = "" },
		"plan bee":            func(p *Payload) { p.Plan = "bee" },
		"an unknown plan":     func(p *Payload) { p.Plan = "swarm" },
		"no installations":    func(p *Payload) { p.Installations = nil },
		"an empty one":        func(p *Payload) { p.Installations = []string{""} },
		"no issuedAt":         func(p *Payload) { p.IssuedAt = time.Time{} },
		"expires when issued": func(p *Payload) { p.ExpiresAt = p.IssuedAt },
		"a negative grace":    func(p *Payload) { p.GraceDays = -1 },
		"grace over a year":   func(p *Payload) { p.GraceDays = maxGraceDays + 1 },
		"a negative limit":    func(p *Payload) { p.Limits["nodes"] = -1 },
	} {
		check(name, with(change), false, false)
	}
	check("a limit name the app does not know", with(func(p *Payload) { p.Limits["node"] = 5 }), false, true)
	check("a sign-in limit", with(func(p *Payload) { p.Limits["sign-in-providers"] = 0 }), false, true)
}

// The constants copied from the app are the app's.
func TestMatchesTheApp(t *testing.T) {
	if licenseContext != license.Context {
		t.Fatalf("licenseContext %q differs from license.Context %q", licenseContext, license.Context)
	}
	if payloadVersion != license.PayloadVersion {
		t.Fatalf("payloadVersion %d differs from license.PayloadVersion %d", payloadVersion, license.PayloadVersion)
	}
	if algEd25519 != releasesig.AlgEd25519 || algMLDSA65 != releasesig.AlgMLDSA65 {
		t.Fatal("algorithm names differ from releasesig")
	}
	if !slices.Equal(requiredAlgorithms, releasesig.RequiredAlgorithms()) {
		t.Fatalf("requiredAlgorithms %v differ from releasesig %v", requiredAlgorithms, releasesig.RequiredAlgorithms())
	}
	var plans []string
	for _, plan := range base.AllLicensePlans {
		if plan != base.LicensePlanBee {
			plans = append(plans, string(plan))
		}
	}
	for _, plan := range plans {
		if _, found := planLimits[plan]; !found {
			t.Fatalf("plan %q has no entry in planLimits", plan)
		}
	}
	if len(planLimits) != len(plans) {
		t.Fatalf("planLimits names %d plans, the app %d (Bee is never in a key)", len(planLimits), len(plans))
	}
	var limits []string
	for _, limit := range base.AllLicenseLimits {
		if limit != base.LicenseLimitSignInProviders {
			limits = append(limits, string(limit))
		}
	}
	if !slices.Equal(keyLimits, limits) {
		t.Fatalf("keyLimits %v differ from the limits a key can set in the app %v", keyLimits, limits)
	}
}

func TestDecodeKeyText(t *testing.T) {
	envelope := `{"payload":"eA==","sha256":"0","signatures":[]}`
	token := base64.StdEncoding.EncodeToString([]byte(envelope))
	for name, text := range map[string]string{
		"a token":                 token,
		"a token without padding": strings.TrimRight(token, "="),
		"a wrapped token":         "  " + token[:10] + "\r\n" + token[10:] + "\n",
		"the envelope's JSON":     envelope,
		"the JSON indented":       "{\n  \"payload\": \"eA==\",\n  \"sha256\": \"0\",\n  \"signatures\": []\n}\n",
	} {
		decoded, err := decodeKeyText(text)
		if err != nil || string(decoded) != envelope {
			t.Fatalf("%s: got %q, %v", name, decoded, err)
		}
	}
	if _, err := decodeKeyText("not a key!"); err == nil {
		t.Fatal("text that is neither base64 nor JSON must be refused")
	}
}

func TestSignRefusesAMisspelledField(t *testing.T) {
	data, err := json.Marshal(mustIssue(t, colonyOptions()))
	if err != nil {
		t.Fatal(err)
	}
	misspelled := strings.Replace(string(data), `"expiresAt"`, `"expiresOn"`, 1)
	if _, err = decodePayload([]byte(misspelled), true); err == nil {
		t.Fatal("sign must refuse a field the payload does not have")
	}
	if _, err = decodePayload([]byte(misspelled), false); err != nil {
		t.Fatalf("verify reads as the app does, passing over what it does not know: %v", err)
	}
}

func mustIssue(t *testing.T, opts *issueOptions) *Payload {
	t.Helper()
	payload, err := issuePayload(opts, testNow)
	if err != nil {
		t.Fatal(err)
	}
	return payload
}
