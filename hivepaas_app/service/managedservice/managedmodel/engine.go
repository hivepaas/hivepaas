// Package managedmodel is what every managed engine is to HivePaaS, apart from
// any one of them. See docs/superpowers/specs/2026-10-17-managed-templates-design.md §3.
package managedmodel

import (
	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/apptemplateservice/templatemodel"
)

// Engine looks after the apps made from the managed templates that name it.
// Each method is added by the task that needs it, so an engine is never asked
// for what it does not yet do.
type Engine interface {
	Name() templatemodel.ManagedEngine
	// APIRange is the engine API versions this engine speaks: the contracts
	// between it and the templates that name it.
	APIRange() (minAPI, maxAPI int)

	// PrepareDoc adds to an app's rendered document what the engine needs built
	// with it - variables, secrets as files, the archive's mount - in the blocks a
	// template may use, so that the spec builder builds and checks them as it
	// does a template's. It refuses a document missing what the engine API needs.
	PrepareDoc(req *PrepareReq) error

	// Configure finishes the settings the document was built into - what no
	// template may say, such as the server's arguments - and returns the
	// engine's own: its state and its scheduled jobs.
	Configure(req *ConfigureReq) (*ConfigureResp, error)

	// Status is a managed app's state, read from its jobs' last runs.
	Status(req *StatusReq) (*Status, error)

	// Reconfigure applies a change of the app's settings to its setting, in
	// place, and to the jobs it made, which it returns as it rewrote them.
	Reconfigure(req *ReconfigureReq) (*ReconfigureResp, error)

	// ArchiveSecrets are the secrets an archive's credentials are kept in, by
	// name, as PrepareDoc puts them in the app; none for an archive without.
	ArchiveSecrets(archive *Archive) map[string]string

	// Verifier is what the engine's verifier app runs.
	Verifier() VerifierSpec

	// RecoveryCheck is how a restored app is asked whether its recovery ended.
	RecoveryCheck() RecoveryCheck
	// FirstBackupJob is the job of a restored app whose run is its first base
	// backup, by id.
	FirstBackupJob(managed *entity.AppManagedSettings) string
	// FinishRestore takes off a restored app what it was restored with, once its
	// first base backup is taken.
	FinishRestore(req *FinishRestoreReq) (*FinishRestoreResp, error)
}
