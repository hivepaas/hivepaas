// Package releasesig verifies the signatures on release.json.
//
// What an installation fetches is release.signed.json: release.json itself,
// base64, together with its signatures. One file rather than release.json plus a
// signature file beside it, because the two are fetched through a CDN that caches
// each on its own - for a few minutes after a release, a separate pair can come
// back as a new file with the old signature, and the update check fails for no
// reason anybody can see. A single file cannot be half-updated.
//
// release.json tells an installation which version exists and which images that
// version runs, and the updater acts on it. It is fetched from the repository,
// so anybody who can write there - or tamper with the file on its way - could
// otherwise point every installation at an image of their choosing. The file is
// signed offline (tools/releasesign) and accepted only when the signatures verify
// against public keys compiled into the binary.
//
// It is signed twice, with ed25519 and with ML-DSA-65, and both are required.
// ed25519 falls to a large enough quantum computer; ML-DSA is built to withstand
// one but is young, as standards and as code. Requiring both means a forger has
// to break both. The pair matters most for installations that never update: the
// requirement a binary ships with is the one it keeps, so it is set to hold up
// if either algorithm does not. A later binary can drop ed25519 once ML-DSA has
// earned it; releases keep carrying both for as long as older binaries need it.
//
// The format here and the one tools/releasesign writes are one format kept in two
// places: the tool is a single stdlib-only file on purpose and cannot import this.
// The tool's tests check that what it signs verifies here.
package releasesig

import (
	"crypto/ed25519"
	"crypto/mldsa"
	"crypto/sha256"
	"crypto/subtle"
	"crypto/x509"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"encoding/pem"
	"errors"
	"fmt"

	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
)

const (
	// Context is the signing context of both algorithms (Ed25519ctx, and the
	// ML-DSA context string), so a signature made with the same key for another
	// purpose is never accepted as a release.json signature.
	Context = "hivepaas-release-v1"

	AlgEd25519 = "ed25519"
	AlgMLDSA65 = "ml-dsa-65"

	// maxSignatures bounds the work an oversized signature file can cause.
	maxSignatures = 16
)

var (
	errNoPublicKeyPEM = errors.New("no PUBLIC KEY PEM block")
	errUnsupportedKey = errors.New("unsupported key")
)

// RequiredAlgorithms lists the algorithms release.json must carry a valid
// signature in, each by a key this binary trusts.
func RequiredAlgorithms() []string {
	return []string{AlgEd25519, AlgMLDSA65}
}

// Envelope is the content of release.signed.json.
//
// Payload is release.json, base64, exactly the bytes that were signed. SHA256 is
// the hash of those bytes, there so a person can compare the envelope with the
// release.json in the commit being released without decoding anything. It is
// checked all the same, so it can never disagree with a payload that was accepted.
type Envelope struct {
	Payload    string  `json:"payload"`
	SHA256     string  `json:"sha256"`
	Signatures []Entry `json:"signatures"`
}

type Entry struct {
	KeyID     string `json:"keyId"`
	Algorithm string `json:"alg"`
	Sig       string `json:"sig"`
}

// PublicKey is a trusted key of one of the supported algorithms.
type PublicKey struct {
	Algorithm string

	ed25519 ed25519.PublicKey
	mldsa   *mldsa.PublicKey
}

// PublicKeys is the set of trusted keys, by key id.
type PublicKeys map[string]*PublicKey

// ParsePublicKeys decodes PKIX PEM public keys by key id, and checks that together
// they can satisfy every required algorithm.
//
// A key that does not decode fails the whole set rather than being skipped: the
// set is compiled in, so a bad entry is a build mistake, and dropping it quietly
// would turn a key rotation into an outage nobody can explain.
func ParsePublicKeys(pemByKeyID map[string][]byte) (PublicKeys, error) {
	keys := make(PublicKeys, len(pemByKeyID))
	covered := map[string]bool{}
	for keyID, content := range pemByKeyID {
		key, err := parsePublicKey(content)
		if err != nil {
			return nil, hperrors.Wrap(hperrors.ErrReleaseSigningKeyInvalid).WithExtraDetail("key %q: %v", keyID, err)
		}
		keys[keyID] = key
		covered[key.Algorithm] = true
	}
	for _, alg := range RequiredAlgorithms() {
		if !covered[alg] {
			return nil, hperrors.Wrap(hperrors.ErrReleaseSigningKeyInvalid).WithExtraDetail("no %s key", alg)
		}
	}
	return keys, nil
}

func parsePublicKey(content []byte) (*PublicKey, error) {
	block, _ := pem.Decode(content)
	if block == nil || block.Type != "PUBLIC KEY" {
		return nil, errNoPublicKeyPEM
	}
	parsed, err := x509.ParsePKIXPublicKey(block.Bytes)
	if err != nil {
		return nil, hperrors.Wrap(err)
	}
	switch key := parsed.(type) {
	case ed25519.PublicKey:
		return &PublicKey{Algorithm: AlgEd25519, ed25519: key}, nil
	case *mldsa.PublicKey:
		if key.Parameters() != mldsa.MLDSA65() {
			return nil, fmt.Errorf("%w: %s", errUnsupportedKey, key.Parameters())
		}
		return &PublicKey{Algorithm: AlgMLDSA65, mldsa: key}, nil
	default:
		return nil, fmt.Errorf("%w: %T", errUnsupportedKey, parsed)
	}
}

// verify reports whether sig is a valid signature over data by k.
func (k *PublicKey) verify(data, sig []byte) bool {
	switch k.Algorithm {
	case AlgEd25519:
		return ed25519.VerifyWithOptions(k.ed25519, data, sig, &ed25519.Options{Context: Context}) == nil
	case AlgMLDSA65:
		return mldsa.Verify(k.mldsa, data, sig, &mldsa.Options{Context: Context}) == nil
	default:
		return false
	}
}

// Open returns the release.json carried by envelope, if the envelope carries,
// for every required algorithm, a valid signature over it by a trusted key of
// that algorithm. Nothing in the payload should be read before Open returns it.
//
// A signature by a key this binary does not know is passed over: it is how a
// release signs with a key that only newer binaries trust, during a rotation or
// once another algorithm is added. A signature by a known key that does not
// verify, or claims a different algorithm than the key has, refuses the whole
// envelope - nothing legitimate produces one.
func Open(keys PublicKeys, envelope []byte) ([]byte, error) {
	invalid := func(format string, args ...any) error {
		return hperrors.Wrap(hperrors.ErrReleaseSignatureInvalid).WithExtraDetail(format, args...)
	}

	var env Envelope
	if err := json.Unmarshal(envelope, &env); err != nil {
		return nil, invalid("malformed envelope")
	}
	if len(env.Signatures) > maxSignatures {
		return nil, invalid("too many signatures")
	}
	data, err := base64.StdEncoding.DecodeString(env.Payload)
	if err != nil || len(data) == 0 {
		return nil, invalid("malformed payload")
	}
	sum := sha256.Sum256(data)
	if subtle.ConstantTimeCompare([]byte(hex.EncodeToString(sum[:])), []byte(env.SHA256)) != 1 {
		return nil, invalid("sha256 does not match")
	}

	satisfied := map[string]bool{}
	for _, entry := range env.Signatures {
		key, found := keys[entry.KeyID]
		if !found {
			continue
		}
		if entry.Algorithm != key.Algorithm {
			return nil, invalid("key %q is %s, signature claims %q", entry.KeyID, key.Algorithm, entry.Algorithm)
		}
		sig, err := base64.StdEncoding.DecodeString(entry.Sig)
		if err != nil || !key.verify(data, sig) {
			return nil, invalid("signature by key %q does not verify", entry.KeyID)
		}
		satisfied[key.Algorithm] = true
	}

	for _, alg := range RequiredAlgorithms() {
		if !satisfied[alg] {
			return nil, invalid("no valid %s signature by a trusted key", alg)
		}
	}
	return data, nil
}
