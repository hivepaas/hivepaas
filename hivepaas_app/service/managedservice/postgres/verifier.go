package postgres

import (
	"github.com/hivepaas/hivepaas/hivepaas_app/service/managedservice/managedmodel"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/managedservice/walgkit"
)

// Verifier idles: walg-entrypoint hands a command that is not postgres to the
// image's entrypoint, which runs it as it is, and walg-restore-test starts the
// server it restores into itself, with archiving off.
func (e *Engine) Verifier() managedmodel.VerifierSpec {
	return managedmodel.VerifierSpec{Command: "sleep infinity", ScratchDir: walgkit.VerifierScratchDir}
}
