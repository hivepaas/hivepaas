package config

type Artifacts struct {
	// Mirror is the base URL of a mirror that holds artifacts by their sha256,
	// <mirror>/<hex>, tried before the URL a template names: for an installation
	// that cannot reach where artifacts are published, or would rather not depend
	// on it. A file from the mirror is checked against its hash like any other.
	Mirror string `toml:"mirror" env:"HP_ARTIFACTS_MIRROR"`
}
