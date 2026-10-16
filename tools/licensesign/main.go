// Command licensesign creates license signing keys and signs license keys with
// them.
//
// A license key lifts the limits of HivePaaS's free plan on the installation it
// names. The portal signs the keys it sells, unattended; this tool signs keys by
// hand: for support, and for development and the e2e environment, with the
// development keys in devkeys/, which only a HivePaaS in development trusts.
//
// A key is the envelope tools/releasesign writes - the payload, base64, with an
// ed25519 and an ML-DSA-65 signature, both required (see
// hivepaas_app/pkg/releasesig) - signed under the context "hivepaas-license-v1".
// The tool knows no other context, so it cannot sign a release.json even when
// handed the release keys.
//
// As text, a key is the envelope's compact JSON as base64 on one line (-format
// token, the default), which survives email and shell quoting. -format json
// writes the envelope with one field per line, for a person to read. The app
// takes either, and ignores whitespace in both.
//
// This file is deliberately one file importing only the standard library, as
// tools/releasesign is, for the same reason: it is handed the production license
// keys to sign a support key, and a short stdlib-only file keeps a change that
// sends them elsewhere visible in review. So it cannot import
// hivepaas_app/pkg/license, and the payload's rules are copied here. Its tests
// check that what it signs verifies there.
//
// A key's id is its file name: <dir>/<key-id>.key and <dir>/<key-id>.pub.pem.
// Keys are PKCS#8 / PKIX PEM.
//
// Usage:
//
//	licensesign keygen -alg ed25519   -key-id lic-2026-ed -dir /secure
//	licensesign keygen -alg ml-dsa-65 -key-id lic-2026-ml -dir /secure
//	licensesign sign   -key A.key -key B.key -in payload.json [-out key.txt] [-format token|json]
//	licensesign issue  -key A.key -key B.key -plan comb|hive|colony|apiary -installation ID [-installation ID]
//	                   [-days 30] [-grace-days 14] [-trial] [-nodes N] [-users N] [-license-id ID]
//	                   [-customer-id ID] [-customer-name NAME] [-customer-email EMAIL] [-subscription-id ID]
//	                   [-out key.txt] [-format token|json]
//	licensesign verify -pub A.pub.pem -pub B.pub.pem [-in key.txt]
//
// sign and issue write the key to standard output unless -out names a file, and
// what they did to standard error, so $(licensesign issue ...) is the key alone.
// A development key, for the swarm of the docker the command talks to:
//
//	licensesign issue -key tools/licensesign/devkeys/dev-ed.key -key tools/licensesign/devkeys/dev-ml.key \
//	                  -plan colony -installation "$(docker info -f '{{.Swarm.Cluster.ID}}')"
package main

import (
	"bytes"
	"crypto"
	"crypto/ed25519"
	"crypto/mldsa"
	"crypto/rand"
	"crypto/sha256"
	"crypto/x509"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"encoding/pem"
	"errors"
	"flag"
	"fmt"
	"io"
	"maps"
	"os"
	"path/filepath"
	"regexp"
	"slices"
	"sort"
	"strings"
	"time"
	"unicode"
)

const (
	// licenseContext must match license.Context.
	licenseContext = "hivepaas-license-v1"
	// payloadVersion must match license.PayloadVersion.
	payloadVersion = 1
	// maxGraceDays must match the bound in license.Payload's validation.
	maxGraceDays = 365

	algEd25519 = "ed25519"
	algMLDSA65 = "ml-dsa-65"

	formatToken = "token"
	formatJSON  = "json"

	// issue's defaults: a month's key, with two weeks' grace.
	defaultDays      = 30
	defaultGraceDays = 14

	// licenseIDBytes is the randomness of a license id issue makes up.
	licenseIDBytes = 8

	day = 24 * time.Hour
)

// requiredAlgorithms must match releasesig.RequiredAlgorithms.
var requiredAlgorithms = []string{algEd25519, algMLDSA65}

// keyIDPattern keeps key ids short and safe to put in file names and JSON.
var keyIDPattern = regexp.MustCompile(`^[A-Za-z0-9][A-Za-z0-9._-]{0,31}$`)

// planLimits are the limits a key of each plan carries unless told otherwise:
// the plan table of the licensing design. A limit left out is no limit. Apiary
// has none of its own: its nodes and users are set per key.
var planLimits = map[string]map[string]int{
	"comb":   {"nodes": 5, "users": 5},
	"hive":   {"nodes": 10, "users": 10},
	"colony": {"nodes": 25, "users": 30},
	"apiary": nil,
}

// keyLimits are the limit names a key can set: the app's, less
// sign-in-providers, which is not a count and which every paid plan lifts. The
// app passes over a name it does not know, so a misspelled one here would be a
// limit that silently is not.
var keyLimits = []string{"nodes", "users", "apps", "functions", "previews"}

// Payload is what a license key says; license.Payload reads the same fields.
type Payload struct {
	Version        int            `json:"version"`
	LicenseID      string         `json:"licenseId"`
	SubscriptionID string         `json:"subscriptionId,omitempty"`
	Customer       Customer       `json:"customer"`
	Plan           string         `json:"plan"`
	Limits         map[string]int `json:"limits,omitempty"`
	Installations  []string       `json:"installations"`
	IssuedAt       time.Time      `json:"issuedAt"`
	ExpiresAt      time.Time      `json:"expiresAt"`
	GraceDays      int            `json:"graceDays"`
	Trial          bool           `json:"trial,omitempty"`
}

type Customer struct {
	ID    string `json:"id,omitempty"`
	Name  string `json:"name,omitempty"`
	Email string `json:"email,omitempty"`
}

// Envelope is a license key: the payload, base64, exactly the bytes signed;
// SHA256 is their hash. The same format as release.signed.json.
type Envelope struct {
	Payload    string     `json:"payload"`
	SHA256     string     `json:"sha256"`
	Signatures []SigEntry `json:"signatures"`
}

type SigEntry struct {
	KeyID     string `json:"keyId"`
	Algorithm string `json:"alg"`
	Sig       string `json:"sig"`
}

type privateKey struct {
	id  string
	alg string
	key crypto.Signer
}

type publicKey struct {
	id  string
	alg string
	key crypto.PublicKey
}

func main() {
	if len(os.Args) < 2 { //nolint:mnd
		usage()
	}
	var err error
	switch os.Args[1] {
	case "keygen":
		err = runKeygen(os.Args[2:])
	case "sign":
		err = runSign(os.Args[2:])
	case "issue":
		err = runIssue(os.Args[2:])
	case "verify":
		err = runVerify(os.Args[2:])
	default:
		usage()
	}
	if err != nil {
		fmt.Fprintln(os.Stderr, "licensesign:", err)
		os.Exit(1)
	}
}

func usage() {
	fmt.Fprintln(os.Stderr, "usage: licensesign keygen|sign|issue|verify [flags]  (-h on a subcommand for its flags)")
	os.Exit(2) //nolint:mnd
}

type multiFlag []string

func (m *multiFlag) String() string     { return strings.Join(*m, ",") }
func (m *multiFlag) Set(v string) error { *m = append(*m, v); return nil }

func runKeygen(args []string) error {
	fs := flag.NewFlagSet("keygen", flag.ExitOnError)
	alg := fs.String("alg", "", "ed25519 or ml-dsa-65")
	keyID := fs.String("key-id", "", "id of the new key, e.g. lic-2026-ed; also its file name")
	dir := fs.String("dir", ".", "directory to write <key-id>.key and <key-id>.pub.pem into")
	_ = fs.Parse(args)

	if !keyIDPattern.MatchString(*keyID) {
		return fmt.Errorf("-key-id %q must match %s", *keyID, keyIDPattern)
	}

	var signer crypto.Signer
	switch *alg {
	case algEd25519:
		_, priv, err := ed25519.GenerateKey(rand.Reader)
		if err != nil {
			return err
		}
		signer = priv
	case algMLDSA65:
		priv, err := mldsa.GenerateKey(mldsa.MLDSA65())
		if err != nil {
			return err
		}
		signer = priv
	default:
		return fmt.Errorf("-alg must be %s or %s", algEd25519, algMLDSA65)
	}

	privDER, err := x509.MarshalPKCS8PrivateKey(signer)
	if err != nil {
		return err
	}
	pubDER, err := x509.MarshalPKIXPublicKey(signer.Public())
	if err != nil {
		return err
	}

	privPath := filepath.Join(*dir, *keyID+".key")
	pubPath := filepath.Join(*dir, *keyID+".pub.pem")
	// O_EXCL: an existing key is never overwritten. Losing the private half of a
	// key installations trust leaves the keys it signed unrenewable until a new
	// binary trusts another.
	if err = writeNew(privPath, pem.EncodeToMemory(&pem.Block{Type: "PRIVATE KEY", Bytes: privDER}), 0o600); err != nil {
		return err
	}
	if err = writeNew(pubPath, pem.EncodeToMemory(&pem.Block{Type: "PUBLIC KEY", Bytes: pubDER}), 0o644); err != nil {
		return err
	}

	fmt.Printf("%s key %q\n", *alg, *keyID)
	fmt.Printf("  private: %s (only where license keys are signed, never in a repository)\n", privPath)
	fmt.Printf("  public:  %s (copy into hivepaas_app/pkg/license/licensekeys/prod/)\n", pubPath)
	return nil
}

func runSign(args []string) error {
	fs := flag.NewFlagSet("sign", flag.ExitOnError)
	var keyFiles multiFlag
	fs.Var(&keyFiles, "key", "private key file <key-id>.key (repeat: one per algorithm)")
	in := fs.String("in", "", "payload to sign, JSON (- for standard input)")
	out := fs.String("out", "-", "file to write the key to (- for standard output)")
	format := fs.String("format", formatToken, "token (base64, one line) or json")
	_ = fs.Parse(args)

	if *in == "" {
		return errors.New("-in is required")
	}
	keys, err := readPrivateKeys(keyFiles)
	if err != nil {
		return err
	}
	content, err := readInput(*in)
	if err != nil {
		return err
	}
	payload, err := decodePayload(content, true)
	if err != nil {
		return fmt.Errorf("%s: %w", *in, err)
	}
	if err = validatePayload(payload); err != nil {
		return fmt.Errorf("%s: %w; refusing to sign it", *in, err)
	}
	// Signed compacted, its fields in the file's order, so the key stays short.
	var data bytes.Buffer
	if err = json.Compact(&data, content); err != nil {
		return err
	}
	return signAndWrite(keys, data.Bytes(), payload, *out, *format)
}

// issueOptions are issue's flags.
type issueOptions struct {
	plan           string
	installations  []string
	days           int
	graceDays      int
	trial          bool
	nodes          int
	users          int
	licenseID      string
	subscriptionID string
	customer       Customer
}

func runIssue(args []string) error {
	fs := flag.NewFlagSet("issue", flag.ExitOnError)
	var keyFiles, installations multiFlag
	fs.Var(&keyFiles, "key", "private key file <key-id>.key (repeat: one per algorithm)")
	fs.Var(&installations, "installation", "swarm cluster ID the key is for (repeat for an Apiary key)")
	opts := &issueOptions{}
	fs.StringVar(&opts.plan, "plan", "", "comb, hive, colony or apiary")
	fs.IntVar(&opts.days, "days", defaultDays, "days the key holds from now, before its grace")
	fs.IntVar(&opts.graceDays, "grace-days", defaultGraceDays, "days the key still holds after it expires")
	fs.BoolVar(&opts.trial, "trial", false, "a trial key, which has no grace")
	fs.IntVar(&opts.nodes, "nodes", -1, "most nodes; -1 is the plan's, and apiary has none")
	fs.IntVar(&opts.users, "users", -1, "most users; -1 is the plan's, and apiary has none")
	fs.StringVar(&opts.licenseID, "license-id", "", "license id (default: a new random one)")
	fs.StringVar(&opts.subscriptionID, "subscription-id", "", "the Paddle subscription the key is for, if any")
	fs.StringVar(&opts.customer.ID, "customer-id", "", "the portal account the key is for")
	fs.StringVar(&opts.customer.Name, "customer-name", "", "the customer's name")
	fs.StringVar(&opts.customer.Email, "customer-email", "", "the customer's email")
	out := fs.String("out", "-", "file to write the key to (- for standard output)")
	format := fs.String("format", formatToken, "token (base64, one line) or json")
	_ = fs.Parse(args)
	opts.installations = installations

	keys, err := readPrivateKeys(keyFiles)
	if err != nil {
		return err
	}
	payload, err := issuePayload(opts, time.Now())
	if err != nil {
		return err
	}
	data, err := json.Marshal(payload)
	if err != nil {
		return err
	}
	return signAndWrite(keys, data, payload, *out, *format)
}

// issuePayload builds the payload issue signs, from its flags, at now.
func issuePayload(opts *issueOptions, now time.Time) (*Payload, error) {
	defaults, known := planLimits[opts.plan]
	if !known {
		return nil, fmt.Errorf("-plan must be comb, hive, colony or apiary, not %q", opts.plan)
	}
	if opts.days < 1 {
		return nil, errors.New("-days must be 1 or more")
	}
	limits := maps.Clone(defaults)
	if limits == nil {
		limits = map[string]int{}
	}
	for _, flagged := range []struct {
		name  string
		value int
	}{{"nodes", opts.nodes}, {"users", opts.users}} {
		switch {
		case flagged.value >= 0:
			limits[flagged.name] = flagged.value
		case defaults == nil:
			return nil, fmt.Errorf("-%s is required: an %s key has no default limits", flagged.name, opts.plan)
		}
	}
	licenseID := opts.licenseID
	if licenseID == "" {
		random := make([]byte, licenseIDBytes)
		if _, err := rand.Read(random); err != nil {
			return nil, err
		}
		licenseID = "lic_" + hex.EncodeToString(random)
	}
	graceDays := opts.graceDays
	if opts.trial {
		// A trial runs out on its date: there is nothing to renew it with.
		graceDays = 0
	}

	issuedAt := now.UTC().Truncate(time.Second)
	payload := &Payload{
		Version:        payloadVersion,
		LicenseID:      licenseID,
		SubscriptionID: opts.subscriptionID,
		Customer:       opts.customer,
		Plan:           opts.plan,
		Limits:         limits,
		Installations:  opts.installations,
		IssuedAt:       issuedAt,
		ExpiresAt:      issuedAt.Add(time.Duration(opts.days) * day),
		GraceDays:      graceDays,
		Trial:          opts.trial,
	}
	if err := validatePayload(payload); err != nil {
		return nil, err
	}
	return payload, nil
}

// signAndWrite signs data, the bytes of payload, opens the result again with the
// public halves of the keys, and writes it out in format.
func signAndWrite(keys []*privateKey, data []byte, payload *Payload, out, format string) error {
	if format != formatToken && format != formatJSON {
		return fmt.Errorf("-format must be %s or %s, not %q", formatToken, formatJSON, format)
	}
	env, err := Sign(keys, data)
	if err != nil {
		return err
	}
	content := FormatKey(env, format)

	// Open what is about to be written, with the public halves of the keys just
	// used, so a key the app would refuse is never handed out.
	pubs := make([]*publicKey, 0, len(keys))
	for _, key := range keys {
		pubs = append(pubs, &publicKey{id: key.id, alg: key.alg, key: key.key.Public()})
	}
	opened, err := Open(pubs, content)
	if err != nil {
		return fmt.Errorf("self-check failed: %w", err)
	}
	if !bytes.Equal(opened, data) {
		return errors.New("self-check failed: the key does not carry the payload that was signed")
	}

	if out == "-" {
		_, err = os.Stdout.Write(content)
	} else {
		// The key carries the customer's email.
		err = os.WriteFile(out, content, 0o600)
	}
	if err != nil {
		return err
	}

	fmt.Fprintf(os.Stderr, "license: %s, %s%s, for %s\n", payload.LicenseID, payload.Plan, trialNote(payload),
		strings.Join(payload.Installations, ", "))
	fmt.Fprintf(os.Stderr, "holds:   %s to %s, then %d grace days\n", payload.IssuedAt.Format(time.RFC3339),
		payload.ExpiresAt.Format(time.RFC3339), payload.GraceDays)
	fmt.Fprintf(os.Stderr, "limits:  %s\n", describeLimits(payload.Limits))
	for _, entry := range env.Signatures {
		fmt.Fprintf(os.Stderr, "key:     %s (%s)\n", entry.KeyID, entry.Algorithm)
	}
	if out != "-" {
		fmt.Fprintf(os.Stderr, "written: %s\n", out)
	}
	return nil
}

func runVerify(args []string) error {
	fs := flag.NewFlagSet("verify", flag.ExitOnError)
	var pubFiles multiFlag
	fs.Var(&pubFiles, "pub", "public key file <key-id>.pub.pem (repeat)")
	in := fs.String("in", "-", "license key, token or JSON (- for standard input)")
	_ = fs.Parse(args)

	pubs := make([]*publicKey, 0, len(pubFiles))
	for _, path := range pubFiles {
		pub, err := readPublicKey(path)
		if err != nil {
			return err
		}
		pubs = append(pubs, pub)
	}
	text, err := readInput(*in)
	if err != nil {
		return err
	}
	data, err := Open(pubs, text)
	if err != nil {
		return err
	}
	// As the app reads it: a field it does not know is passed over.
	payload, err := decodePayload(data, false)
	if err != nil {
		return err
	}
	if err = validatePayload(payload); err != nil {
		return fmt.Errorf("validly signed, but the app would refuse it: %w", err)
	}
	pretty, err := json.MarshalIndent(payload, "", "  ")
	if err != nil {
		return err
	}
	fmt.Printf("OK: valid %s signatures, context %s\n%s\n", strings.Join(requiredAlgorithms, " and "),
		licenseContext, pretty)
	return nil
}

// decodePayload reads a payload. strict refuses a field the payload does not
// have, which in a payload written by hand is a misspelling.
func decodePayload(data []byte, strict bool) (*Payload, error) {
	decoder := json.NewDecoder(bytes.NewReader(data))
	if strict {
		decoder.DisallowUnknownFields()
	}
	var payload Payload
	if err := decoder.Decode(&payload); err != nil {
		return nil, fmt.Errorf("malformed payload: %w", err)
	}
	return &payload, nil
}

// validatePayload holds a payload to the rules license.Payload's validation
// applies, and refuses limit names a key cannot set.
func validatePayload(p *Payload) error {
	if p.Version != payloadVersion {
		return fmt.Errorf("version must be %d", payloadVersion)
	}
	if p.LicenseID == "" {
		return errors.New("licenseId is required")
	}
	if _, known := planLimits[p.Plan]; !known {
		return fmt.Errorf("plan must be comb, hive, colony or apiary, not %q", p.Plan)
	}
	if len(p.Installations) == 0 {
		return errors.New("at least one installation is required")
	}
	if slices.Contains(p.Installations, "") {
		return errors.New("an installation ID is empty")
	}
	if p.IssuedAt.IsZero() || p.ExpiresAt.IsZero() {
		return errors.New("issuedAt and expiresAt are both required")
	}
	if !p.ExpiresAt.After(p.IssuedAt) {
		return errors.New("expiresAt must be after issuedAt")
	}
	if p.GraceDays < 0 || p.GraceDays > maxGraceDays {
		return fmt.Errorf("graceDays must be between 0 and %d", maxGraceDays)
	}
	for name, value := range p.Limits {
		if !slices.Contains(keyLimits, name) {
			return fmt.Errorf("limit %q is not one a key can set (%s)", name, strings.Join(keyLimits, ", "))
		}
		if value < 0 {
			return fmt.Errorf("limit %q must not be negative", name)
		}
	}
	return nil
}

// Sign signs data under the license context with every key, and refuses unless
// the keys cover every required algorithm exactly once.
func Sign(keys []*privateKey, data []byte) (*Envelope, error) {
	seen := map[string]string{}
	for _, key := range keys {
		if other, dup := seen[key.alg]; dup {
			return nil, fmt.Errorf("keys %q and %q are both %s; give one key per algorithm", other, key.id, key.alg)
		}
		seen[key.alg] = key.id
	}
	for _, alg := range requiredAlgorithms {
		if _, found := seen[alg]; !found {
			return nil, fmt.Errorf("no %s key given; a license key needs a signature in each of %v",
				alg, requiredAlgorithms)
		}
	}

	sum := sha256.Sum256(data)
	env := &Envelope{
		Payload: base64.StdEncoding.EncodeToString(data),
		SHA256:  hex.EncodeToString(sum[:]),
	}
	for _, key := range keys {
		var sig []byte
		var err error
		switch k := key.key.(type) {
		case ed25519.PrivateKey:
			sig, err = k.Sign(nil, data, &ed25519.Options{Context: licenseContext})
		case *mldsa.PrivateKey:
			// Deterministic, so signing the same payload again gives the same key.
			sig, err = k.SignDeterministic(data, &mldsa.Options{Context: licenseContext})
		default:
			err = fmt.Errorf("key %q: unsupported key type %T", key.id, key.key)
		}
		if err != nil {
			return nil, err
		}
		env.Signatures = append(env.Signatures, SigEntry{
			KeyID:     key.id,
			Algorithm: key.alg,
			Sig:       base64.StdEncoding.EncodeToString(sig),
		})
	}
	sort.Slice(env.Signatures, func(i, j int) bool {
		return env.Signatures[i].KeyID < env.Signatures[j].KeyID
	})
	return env, nil
}

// Open applies the app's rule and returns the payload: for every required
// algorithm, a valid signature under the license context by one of pubs. A
// signature by a key not in pubs is passed over; one by a key in pubs that does
// not verify, or names the wrong algorithm, fails the envelope. content is a
// key's text, either form.
func Open(pubs []*publicKey, content []byte) ([]byte, error) {
	envelope, err := decodeKeyText(string(content))
	if err != nil {
		return nil, err
	}
	var env Envelope
	if err = json.Unmarshal(envelope, &env); err != nil {
		return nil, fmt.Errorf("malformed envelope: %w", err)
	}
	data, err := base64.StdEncoding.DecodeString(env.Payload)
	if err != nil || len(data) == 0 {
		return nil, errors.New("malformed payload")
	}
	sum := sha256.Sum256(data)
	if env.SHA256 != hex.EncodeToString(sum[:]) {
		return nil, errors.New("sha256 does not match the payload")
	}

	byID := map[string]*publicKey{}
	for _, pub := range pubs {
		byID[pub.id] = pub
	}
	satisfied := map[string]bool{}
	for _, entry := range env.Signatures {
		pub, found := byID[entry.KeyID]
		if !found {
			continue
		}
		if entry.Algorithm != pub.alg {
			return nil, fmt.Errorf("key %q is %s, signature claims %q", entry.KeyID, pub.alg, entry.Algorithm)
		}
		sig, err := base64.StdEncoding.DecodeString(entry.Sig)
		if err != nil {
			return nil, fmt.Errorf("signature by key %q is malformed", entry.KeyID)
		}
		switch k := pub.key.(type) {
		case ed25519.PublicKey:
			err = ed25519.VerifyWithOptions(k, data, sig, &ed25519.Options{Context: licenseContext})
		case *mldsa.PublicKey:
			err = mldsa.Verify(k, data, sig, &mldsa.Options{Context: licenseContext})
		default:
			err = fmt.Errorf("unsupported key type %T", pub.key)
		}
		if err != nil {
			return nil, fmt.Errorf("signature by key %q does not verify: %w", entry.KeyID, err)
		}
		satisfied[pub.alg] = true
	}
	for _, alg := range requiredAlgorithms {
		if !satisfied[alg] {
			return nil, fmt.Errorf("no valid %s signature by the given keys", alg)
		}
	}
	return data, nil
}

// FormatKey writes a key's text. A token is the envelope's compact JSON as
// base64, on one line with a newline after it; json is the envelope with each
// field on a line of its own.
func FormatKey(env *Envelope, format string) []byte {
	if format == formatJSON {
		var b strings.Builder
		fmt.Fprintf(&b, "{\n  \"payload\": %q,\n  \"sha256\": %q,\n  \"signatures\": [\n", env.Payload, env.SHA256)
		for i, entry := range env.Signatures {
			line, _ := json.Marshal(entry)
			b.WriteString("    ")
			b.Write(line)
			if i < len(env.Signatures)-1 {
				b.WriteString(",")
			}
			b.WriteString("\n")
		}
		b.WriteString("  ]\n}\n")
		return []byte(b.String())
	}
	compact, _ := json.Marshal(env)
	return []byte(base64.StdEncoding.EncodeToString(compact) + "\n")
}

// decodeKeyText is license.DecodeKeyText: whitespace is removed, then the text
// is the envelope's JSON, or that JSON as base64 with or without padding.
func decodeKeyText(text string) ([]byte, error) {
	text = strings.Map(func(r rune) rune {
		if unicode.IsSpace(r) {
			return -1
		}
		return r
	}, text)
	if strings.HasPrefix(text, "{") {
		return []byte(text), nil
	}
	envelope, err := base64.RawStdEncoding.DecodeString(strings.TrimRight(text, "="))
	if err != nil || len(envelope) == 0 {
		return nil, errors.New("not a license key: neither base64 nor JSON")
	}
	return envelope, nil
}

func trialNote(p *Payload) string {
	if p.Trial {
		return " (trial)"
	}
	return ""
}

func describeLimits(limits map[string]int) string {
	if len(limits) == 0 {
		return "none"
	}
	parts := make([]string, 0, len(limits))
	for _, name := range keyLimits {
		if value, found := limits[name]; found {
			parts = append(parts, fmt.Sprintf("%d %s", value, name))
		}
	}
	return strings.Join(parts, ", ")
}

func readInput(path string) ([]byte, error) {
	if path == "-" {
		return io.ReadAll(os.Stdin)
	}
	return os.ReadFile(path)
}

func readPrivateKeys(paths []string) ([]*privateKey, error) {
	keys := make([]*privateKey, 0, len(paths))
	for _, path := range paths {
		key, err := readPrivateKey(path)
		if err != nil {
			return nil, err
		}
		keys = append(keys, key)
	}
	return keys, nil
}

func readPrivateKey(path string) (*privateKey, error) {
	id, err := keyIDFromPath(path, ".key")
	if err != nil {
		return nil, err
	}
	block, err := readPEM(path, "PRIVATE KEY")
	if err != nil {
		return nil, err
	}
	parsed, err := x509.ParsePKCS8PrivateKey(block.Bytes)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", path, err)
	}
	switch k := parsed.(type) {
	case ed25519.PrivateKey:
		return &privateKey{id: id, alg: algEd25519, key: k}, nil
	case *mldsa.PrivateKey:
		if k.PublicKey().Parameters() != mldsa.MLDSA65() {
			return nil, fmt.Errorf("%s: ML-DSA parameter set %s is not supported", path, k.PublicKey().Parameters())
		}
		return &privateKey{id: id, alg: algMLDSA65, key: k}, nil
	default:
		return nil, fmt.Errorf("%s: unsupported key type %T", path, parsed)
	}
}

func readPublicKey(path string) (*publicKey, error) {
	id, err := keyIDFromPath(path, ".pub.pem")
	if err != nil {
		return nil, err
	}
	block, err := readPEM(path, "PUBLIC KEY")
	if err != nil {
		return nil, err
	}
	parsed, err := x509.ParsePKIXPublicKey(block.Bytes)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", path, err)
	}
	switch k := parsed.(type) {
	case ed25519.PublicKey:
		return &publicKey{id: id, alg: algEd25519, key: k}, nil
	case *mldsa.PublicKey:
		if k.Parameters() != mldsa.MLDSA65() {
			return nil, fmt.Errorf("%s: ML-DSA parameter set %s is not supported", path, k.Parameters())
		}
		return &publicKey{id: id, alg: algMLDSA65, key: k}, nil
	default:
		return nil, fmt.Errorf("%s: unsupported key type %T", path, parsed)
	}
}

func keyIDFromPath(path, suffix string) (string, error) {
	base := filepath.Base(path)
	id, found := strings.CutSuffix(base, suffix)
	if !found || !keyIDPattern.MatchString(id) {
		return "", fmt.Errorf("%s: key files are named <key-id>%s, with a key id matching %s", path, suffix, keyIDPattern)
	}
	return id, nil
}

func readPEM(path, wantType string) (*pem.Block, error) {
	content, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	block, _ := pem.Decode(content)
	if block == nil || block.Type != wantType {
		return nil, fmt.Errorf("%s: no %s PEM block", path, wantType)
	}
	return block, nil
}

func writeNew(path string, content []byte, perm os.FileMode) error {
	f, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_EXCL, perm)
	if err != nil {
		return err
	}
	if _, err = f.Write(content); err != nil {
		_ = f.Close()
		return err
	}
	return f.Close()
}
