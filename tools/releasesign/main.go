// Command releasesign creates the signing keys for release.json and signs
// release.json with them.
//
// An installation learns that a newer HivePaaS exists, and which images that
// version runs, from release.json. The signatures let it tell a file we released
// from one that anybody with write access to the repository - or to anything
// between the repository and the installation - put there.
//
// release.json is signed twice, with ed25519 and with ML-DSA-65, and the app
// requires both (see hivepaas_app/pkg/releasesig for why). `sign` therefore
// refuses to write a signature file that lacks either.
//
// This file is deliberately one file importing only the standard library, and it
// must stay that way. `make release-sign` builds it from a pinned commit outside
// any module with the network off, so an import from outside the standard library
// fails that build rather than being fetched. A tool that holds the private keys
// is the obvious place to plant something that sends them elsewhere; keeping it
// to one short, stdlib-only file keeps such a change visible in review.
//
// Both algorithms sign with the context "hivepaas-release-v1" (Ed25519ctx, and
// the ML-DSA context string). The context ties a signature to this purpose: a key
// that also signed something else could not have that passed off as a
// release.json signature. The verifier in the app must use the same context.
//
// A key's id is its file name: <dir>/<key-id>.key and <dir>/<key-id>.pub.pem.
// Keys are PKCS#8 / PKIX PEM, so openssl (3.5 or later, for ML-DSA) reads them as
// well, and `make release-sign` checks every signature with openssl too.
//
// Usage:
//
//	releasesign keygen -alg ed25519   -key-id 2026-ed -dir /offline
//	releasesign keygen -alg ml-dsa-65 -key-id 2026-ml -dir /offline
//	releasesign sign   -key /offline/2026-ed.key -key /offline/2026-ml.key [-in release.json] [-out release.json.sig]
//	releasesign verify -pub 2026-ed.pub.pem -pub 2026-ml.pub.pem [-in release.json] [-sig release.json.sig]
package main

import (
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
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
)

const (
	signContext = "hivepaas-release-v1"

	algEd25519 = "ed25519"
	algMLDSA65 = "ml-dsa-65"
)

// requiredAlgorithms must match releasesig.RequiredAlgorithms.
var requiredAlgorithms = []string{algEd25519, algMLDSA65}

// keyIDPattern keeps key ids short and safe to put in file names and JSON.
var keyIDPattern = regexp.MustCompile(`^[A-Za-z0-9][A-Za-z0-9._-]{0,31}$`)

// SigFile is the content of release.json.sig. SHA256 is for a person to compare
// against the commit being released; the signatures cover the file itself.
type SigFile struct {
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
	case "verify":
		err = runVerify(os.Args[2:])
	default:
		usage()
	}
	if err != nil {
		fmt.Fprintln(os.Stderr, "releasesign:", err)
		os.Exit(1)
	}
}

func usage() {
	fmt.Fprintln(os.Stderr, "usage: releasesign keygen|sign|verify [flags]  (-h on a subcommand for its flags)")
	os.Exit(2) //nolint:mnd
}

type multiFlag []string

func (m *multiFlag) String() string     { return strings.Join(*m, ",") }
func (m *multiFlag) Set(v string) error { *m = append(*m, v); return nil }

func runKeygen(args []string) error {
	fs := flag.NewFlagSet("keygen", flag.ExitOnError)
	alg := fs.String("alg", "", "ed25519 or ml-dsa-65")
	keyID := fs.String("key-id", "", "id of the new key, e.g. 2026-ed; also its file name")
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
	// key installations trust is recoverable only by shipping a new binary.
	if err = writeNew(privPath, pem.EncodeToMemory(&pem.Block{Type: "PRIVATE KEY", Bytes: privDER}), 0o600); err != nil {
		return err
	}
	if err = writeNew(pubPath, pem.EncodeToMemory(&pem.Block{Type: "PUBLIC KEY", Bytes: pubDER}), 0o644); err != nil {
		return err
	}

	fmt.Printf("%s key %q\n", *alg, *keyID)
	fmt.Printf("  private: %s (keep offline, never commit)\n", privPath)
	fmt.Printf("  public:  %s (copy into hivepaas_app/service/hpappservice/hpappserviceimpl/releasekeys/)\n",
		pubPath)
	return nil
}

func runSign(args []string) error {
	fs := flag.NewFlagSet("sign", flag.ExitOnError)
	var keyFiles multiFlag
	fs.Var(&keyFiles, "key", "private key file <key-id>.key (repeat: one per algorithm)")
	in := fs.String("in", "release.json", "file to sign")
	out := fs.String("out", "", "signature file to write (default <in>.sig)")
	_ = fs.Parse(args)

	if *out == "" {
		*out = *in + ".sig"
	}

	keys := make([]*privateKey, 0, len(keyFiles))
	for _, path := range keyFiles {
		key, err := readPrivateKey(path)
		if err != nil {
			return err
		}
		keys = append(keys, key)
	}
	data, err := os.ReadFile(*in)
	if err != nil {
		return err
	}
	if !json.Valid(data) {
		return fmt.Errorf("%s is not valid JSON; refusing to sign it", *in)
	}

	sigFile, err := Sign(keys, data)
	if err != nil {
		return err
	}

	// Check the signatures before writing them, against the public halves of the
	// keys just used, so a file the app would refuse never reaches disk.
	pubs := make([]*publicKey, 0, len(keys))
	for _, key := range keys {
		pubs = append(pubs, &publicKey{id: key.id, alg: key.alg, key: key.key.Public()})
	}
	if err = Verify(pubs, data, sigFile); err != nil {
		return fmt.Errorf("self-check failed: %w", err)
	}

	if err = os.WriteFile(*out, FormatSigFile(sigFile), 0o644); err != nil { //nolint:gosec
		return err
	}

	fmt.Printf("signed:  %s\n", *in)
	fmt.Printf("sha256:  %s  <- compare with `git show <release-ref>:%s | shasum -a 256`\n", sigFile.SHA256, *in)
	for _, entry := range sigFile.Signatures {
		fmt.Printf("key:     %s (%s)\n", entry.KeyID, entry.Algorithm)
	}
	fmt.Printf("written: %s\n", *out)
	return nil
}

func runVerify(args []string) error {
	fs := flag.NewFlagSet("verify", flag.ExitOnError)
	var pubFiles multiFlag
	fs.Var(&pubFiles, "pub", "public key file <key-id>.pub.pem (repeat)")
	in := fs.String("in", "release.json", "signed file")
	sigPath := fs.String("sig", "", "signature file (default <in>.sig)")
	_ = fs.Parse(args)

	if *sigPath == "" {
		*sigPath = *in + ".sig"
	}
	pubs := make([]*publicKey, 0, len(pubFiles))
	for _, path := range pubFiles {
		pub, err := readPublicKey(path)
		if err != nil {
			return err
		}
		pubs = append(pubs, pub)
	}
	data, err := os.ReadFile(*in)
	if err != nil {
		return err
	}
	content, err := os.ReadFile(*sigPath)
	if err != nil {
		return err
	}
	var sigFile SigFile
	if err = json.Unmarshal(content, &sigFile); err != nil {
		return fmt.Errorf("%s: %w", *sigPath, err)
	}
	if err = Verify(pubs, data, &sigFile); err != nil {
		return err
	}
	fmt.Printf("OK: %s carries valid %s signatures\n", *in, strings.Join(requiredAlgorithms, " and "))
	return nil
}

// Sign signs data with every key, and refuses unless the keys cover every
// required algorithm exactly once.
func Sign(keys []*privateKey, data []byte) (*SigFile, error) {
	seen := map[string]string{}
	for _, key := range keys {
		if other, dup := seen[key.alg]; dup {
			return nil, fmt.Errorf("keys %q and %q are both %s; give one key per algorithm", other, key.id, key.alg)
		}
		seen[key.alg] = key.id
	}
	for _, alg := range requiredAlgorithms {
		if _, found := seen[alg]; !found {
			return nil, fmt.Errorf("no %s key given; release.json needs a signature in each of %v",
				alg, requiredAlgorithms)
		}
	}

	sum := sha256.Sum256(data)
	sigFile := &SigFile{SHA256: hex.EncodeToString(sum[:])}
	for _, key := range keys {
		var sig []byte
		var err error
		switch k := key.key.(type) {
		case ed25519.PrivateKey:
			sig, err = k.Sign(nil, data, &ed25519.Options{Context: signContext})
		case *mldsa.PrivateKey:
			// Deterministic, so signing the same file again gives the same bytes.
			sig, err = k.SignDeterministic(data, &mldsa.Options{Context: signContext})
		default:
			err = fmt.Errorf("key %q: unsupported key type %T", key.id, key.key)
		}
		if err != nil {
			return nil, err
		}
		sigFile.Signatures = append(sigFile.Signatures, SigEntry{
			KeyID:     key.id,
			Algorithm: key.alg,
			Sig:       base64.StdEncoding.EncodeToString(sig),
		})
	}
	sort.Slice(sigFile.Signatures, func(i, j int) bool {
		return sigFile.Signatures[i].KeyID < sigFile.Signatures[j].KeyID
	})
	return sigFile, nil
}

// Verify applies the app's rule: for every required algorithm, a valid signature
// by one of pubs. A signature by a key not in pubs is passed over; one by a key
// in pubs that does not verify, or names the wrong algorithm, fails the file.
func Verify(pubs []*publicKey, data []byte, sigFile *SigFile) error {
	byID := map[string]*publicKey{}
	for _, pub := range pubs {
		byID[pub.id] = pub
	}
	sum := sha256.Sum256(data)
	if sigFile.SHA256 != hex.EncodeToString(sum[:]) {
		return errors.New("sha256 does not match the file")
	}

	satisfied := map[string]bool{}
	for _, entry := range sigFile.Signatures {
		pub, found := byID[entry.KeyID]
		if !found {
			continue
		}
		if entry.Algorithm != pub.alg {
			return fmt.Errorf("key %q is %s, signature claims %q", entry.KeyID, pub.alg, entry.Algorithm)
		}
		sig, err := base64.StdEncoding.DecodeString(entry.Sig)
		if err != nil {
			return fmt.Errorf("signature by key %q is malformed", entry.KeyID)
		}
		switch k := pub.key.(type) {
		case ed25519.PublicKey:
			err = ed25519.VerifyWithOptions(k, data, sig, &ed25519.Options{Context: signContext})
		case *mldsa.PublicKey:
			err = mldsa.Verify(k, data, sig, &mldsa.Options{Context: signContext})
		default:
			err = fmt.Errorf("unsupported key type %T", pub.key)
		}
		if err != nil {
			return fmt.Errorf("signature by key %q does not verify: %w", entry.KeyID, err)
		}
		satisfied[pub.alg] = true
	}
	for _, alg := range requiredAlgorithms {
		if !satisfied[alg] {
			return fmt.Errorf("no valid %s signature by the given keys", alg)
		}
	}
	return nil
}

// FormatSigFile writes one signature per line, so a diff of release.json.sig
// shows which key's signature changed and scripts can pick one out by key id.
func FormatSigFile(sigFile *SigFile) []byte {
	var b strings.Builder
	fmt.Fprintf(&b, "{\n  \"sha256\": %q,\n  \"signatures\": [\n", sigFile.SHA256)
	for i, entry := range sigFile.Signatures {
		line, _ := json.Marshal(entry)
		b.WriteString("    ")
		b.Write(line)
		if i < len(sigFile.Signatures)-1 {
			b.WriteString(",")
		}
		b.WriteString("\n")
	}
	b.WriteString("  ]\n}\n")
	return []byte(b.String())
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
