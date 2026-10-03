// Package hivepaas holds the release file at the repository's root for the code:
// the binary is built with the release it is part of.
package hivepaas

import _ "embed"

// ReleaseJSON is release.json as it was when the binary was built: what each
// channel's release runs. base reads its own release from it.
//
//go:embed release.json
var ReleaseJSON []byte
