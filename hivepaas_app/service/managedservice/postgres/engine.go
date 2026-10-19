// Package postgres is the managed engine of Postgres: WAL-G archiving, base
// backups and their retention, restore tests, and restores to a point in time.
package postgres

import (
	"github.com/hivepaas/hivepaas/hivepaas_app/service/apptemplateservice/templatemodel"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/managedservice/managedmodel"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/managedservice/walgkit"
)

const (
	minAPI = 1
	maxAPI = 1
)

// profile is Postgres, as what it shares with the other WAL-G engines needs it.
var profile = &walgkit.Profile{
	Engine:             templatemodel.ManagedEnginePostgres,
	API:                maxAPI,
	Scripts:            apiScripts,
	FileOwner:          postgresUID,
	BackupPattern:      backupPattern,
	DeltaMarker:        "_D_",
	PendingBytesOutput: outPgWalBytes,
	Dump:               walgkit.DumpCommands{Source: dumpCommand, Restore: dumpRestoreCommand, FileExt: ".dump"},
	Members:            membersStatus,
}

type Engine struct{}

func New() *Engine {
	return &Engine{}
}

func (e *Engine) Name() templatemodel.ManagedEngine {
	return templatemodel.ManagedEnginePostgres
}

func (e *Engine) APIRange() (int, int) {
	return minAPI, maxAPI
}

// Capabilities are Postgres's: it has replicas, and runs its restore test in a
// verifier app too.
func (e *Engine) Capabilities() managedmodel.Capabilities {
	return managedmodel.Capabilities{Replicas: true, VerifierApp: true}
}

// PointNameMax is what Postgres keeps of a restore point's name: 63 bytes.
func (e *Engine) PointNameMax() int {
	return pointNameMax
}
