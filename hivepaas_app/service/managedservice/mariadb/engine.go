// Package mariadb is the managed engine of MariaDB: WAL-G archiving of the
// binary logs by the template's archive loop, base backups with mariadb-backup
// and their retention, restore tests, and restores to a point in time by a
// replay of the binary logs. See
// docs/superpowers/specs/2026-10-19-managed-mariadb-design.md.
//
// The engine is built a task at a time (the MariaDB phase 1 plan): what it does
// not do yet, it refuses, so that no app is made, nor a job run, with half an
// engine.
package mariadb

import (
	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/apptemplateservice/templatemodel"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/managedservice/managedmodel"
)

const (
	minAPI = 1
	maxAPI = 1
	// pointNameMax is the restore point names the engine takes: the name is the
	// engine's alone, since MariaDB keeps none, and Postgres's limit keeps one
	// set of names good for both.
	pointNameMax = 63
)

type Engine struct{}

func New() *Engine {
	return &Engine{}
}

func (e *Engine) Name() templatemodel.ManagedEngine {
	return templatemodel.ManagedEngineMariaDB
}

func (e *Engine) APIRange() (int, int) {
	return minAPI, maxAPI
}

// Capabilities are MariaDB's: no replicas yet, and its restore test runs in a
// verifier app too.
func (e *Engine) Capabilities() managedmodel.Capabilities {
	return managedmodel.Capabilities{Replicas: false, VerifierApp: true}
}

func (e *Engine) PointNameMax() int {
	return pointNameMax
}

// unsupported refuses what the engine does not do (yet), named as what a MariaDB
// database does not have.
func unsupported(what string) error {
	return hperrors.Wrap(hperrors.ErrAppManagedUnsupportedByEngine).
		WithParam("Engine", templatemodel.ManagedEngines[templatemodel.ManagedEngineMariaDB].Title).
		WithParam("What", what).ExposeParams("Engine", "What")
}

func (e *Engine) PrepareDoc(*managedmodel.PrepareReq) error {
	return unsupported("archiving")
}

func (e *Engine) Configure(*managedmodel.ConfigureReq) (*managedmodel.ConfigureResp, error) {
	return nil, unsupported("archiving")
}

func (e *Engine) Status(*managedmodel.StatusReq) (*managedmodel.Status, error) {
	return nil, unsupported("a status")
}

func (e *Engine) Reconfigure(*managedmodel.ReconfigureReq) (*managedmodel.ReconfigureResp, error) {
	return nil, unsupported("settings")
}

func (e *Engine) ArchiveSecrets(*managedmodel.Archive) map[string]string {
	return nil
}

func (e *Engine) Verifier() managedmodel.VerifierSpec {
	return managedmodel.VerifierSpec{}
}

func (e *Engine) RecoveryCheck() managedmodel.RecoveryCheck {
	return managedmodel.RecoveryCheck{}
}

func (e *Engine) FirstBackupJob(*entity.AppManagedSettings) string {
	return ""
}

func (e *Engine) FinishRestore(*managedmodel.FinishRestoreReq) (*managedmodel.FinishRestoreResp, error) {
	return nil, unsupported("restores")
}

func (e *Engine) NightlyBackupArgs() string {
	return ""
}

func (e *Engine) Retention(*managedmodel.RetentionReq) (*managedmodel.Retention, error) {
	return nil, unsupported("nightly backups")
}

func (e *Engine) MarkRestorePoint(*entity.AppManagedSettings, string) (managedmodel.MarkCommand, error) {
	return managedmodel.MarkCommand{}, unsupported("restore points")
}

func (e *Engine) MarkResult(string) (string, error) {
	return "", unsupported("restore points")
}

func (e *Engine) StaleDeployPoints(*entity.AppManagedSettings, map[string]string) []string {
	return nil
}

func (e *Engine) Replica(*managedmodel.ReplicaReq) (*managedmodel.ReplicaSpec, error) {
	return nil, unsupported("replicas")
}

func (e *Engine) Promote(*managedmodel.PromoteReq) (*managedmodel.PromoteSpec, error) {
	return nil, unsupported("replicas")
}
