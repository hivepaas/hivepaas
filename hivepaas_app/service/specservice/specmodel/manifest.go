// Package specmodel holds the spec format itself - the manifest, the report and
// the types every payload file is built from.
//
// It is the durable half of the feature. A published bundle is read back by
// versions of HivePaaS that do not exist yet, so the shape here changes only
// with apiVersion, and it deliberately shares no types with the dashboard DTOs,
// which are free to move whenever the UI does.
package specmodel

import "time"

// APIVersion is the format version every bundle and payload file carries.
const APIVersion = "hivepaas.com/v1"

// Kind distinguishes a configuration-only bundle from anything added later - a
// snapshot carrying volume data, for one.
type Kind string

const KindSpec Kind = "Spec"

// SecretsMode decides what happens to the values settings keep encrypted.
type SecretsMode string

const (
	// SecretsModeNone omits every secret. It needs no capability and leaks
	// nothing, which is why it is the default.
	SecretsModeNone SecretsMode = "none"
	// SecretsModeEncrypted includes secrets and wraps the whole bundle with age
	// under a passphrase the operator supplies.
	SecretsModeEncrypted SecretsMode = "encrypted"
	// SecretsModePlaintext includes secrets in the clear.
	SecretsModePlaintext SecretsMode = "plaintext"
)

func (m SecretsMode) IsValid() bool {
	switch m {
	case SecretsModeNone, SecretsModeEncrypted, SecretsModePlaintext:
		return true
	default:
		return false
	}
}

// RevealsSecrets reports whether this mode decrypts anything, and therefore
// whether the export has to pass the secret-reveal capability gate.
//
// Both secret-bearing modes decrypt exactly the same values. The difference
// between them is only whether the plaintext lands on disk or inside an age
// envelope, not whether it was read - so both are gated and audited alike.
func (m SecretsMode) RevealsSecrets() bool {
	return m == SecretsModeEncrypted || m == SecretsModePlaintext
}

// Manifest is spec.yaml at the root of every bundle.
type Manifest struct {
	APIVersion string    `yaml:"apiVersion"`
	Kind       Kind      `yaml:"kind"`
	ExportedAt time.Time `yaml:"exportedAt"`
	// SourceAppVersion is informational - which build produced the bundle.
	SourceAppVersion string `yaml:"sourceAppVersion"`
	// SourceVersionCode is what compatibility is judged on.
	SourceVersionCode string      `yaml:"sourceVersionCode"`
	Scope             string      `yaml:"scope"`
	SecretsMode       SecretsMode `yaml:"secretsMode"`
	// Files is every payload file in the bundle, in bundle order. It is also the
	// surface a partial import selects from.
	Files []string `yaml:"files"`
}
