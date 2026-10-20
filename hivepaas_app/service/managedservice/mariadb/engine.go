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
	"regexp"

	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/apptemplateservice/templatemodel"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/managedservice/managedmodel"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/managedservice/walgkit"
)

const (
	minAPI = 1
	maxAPI = 1
	// mysqlUID owns the secret files: the official images' mysql user, which the
	// archive loop runs as.
	mysqlUID = "999"
	// pointNameMax is the restore point names the engine takes: the name is the
	// engine's alone, since MariaDB keeps none, and Postgres's limit keeps one
	// set of names good for both.
	pointNameMax = walgkit.PointNameMax
)

// profile is MariaDB, as what it shares with the other WAL-G engines needs it.
var profile = &walgkit.Profile{
	Engine:             templatemodel.ManagedEngineMariaDB,
	API:                maxAPI,
	Scripts:            apiScripts,
	FileOwner:          mysqlUID,
	BackupPattern:      backupPattern,
	PendingBytesOutput: outPendingBytes,
	Dump:               walgkit.DumpCommands{Source: dumpCommand, Restore: dumpRestoreCommand, FileExt: ".sql"},
	RestoreTarget:      targetEnv,
	Members:            walgkit.MembersStatus,
	// walg-entrypoint promotes: a template has at most ten config files. Every
	// template whose scripts let replicas in has it.
	Promotion: walgkit.Promotion{Command: walgkit.ScriptDir + "/walg-entrypoint promote",
		Requires: walgkit.ReplicationScript},
	// MariaDB's clients know one host: the cluster's alias reaches its writer.
	WriterAlias: true,
}

// apiScripts are the scripts engine API 1 runs, which the template ships as
// config files in walgkit.ScriptDir; walg-lib is what the others source. What
// each takes, prints and exits with is the contract: the scripts' own comments
// say it, and the MariaDB design §5.
var apiScripts = []string{
	"walg-lib", "walg-run", "walg-archive", "walg-replay", "walg-nightly", "walg-health", "walg-mark",
	"walg-restore-test", "walg-entrypoint",
}

// backupPattern is a base backup's name: WAL-G names MySQL's by the time they
// start, and they are full ones.
var backupPattern = regexp.MustCompile(`^stream_[0-9]{8}T[0-9]{6}Z$`)

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

// Capabilities are MariaDB's: it has replicas, and runs its restore test in a
// verifier app too.
func (e *Engine) Capabilities() managedmodel.Capabilities {
	return managedmodel.Capabilities{Replicas: true, VerifierApp: true}
}

func (e *Engine) PointNameMax() int {
	return pointNameMax
}

// Verifier idles: walg-entrypoint hands a command that is not mariadbd to the
// image's entrypoint, which runs it as it is, and walg-restore-test starts the
// server it restores into itself, with no binary log. It reaches the live
// database as root, its password the database's, in an option file in memory.
func (e *Engine) Verifier() managedmodel.VerifierSpec {
	return managedmodel.VerifierSpec{Command: "sleep infinity", ScratchDir: walgkit.VerifierScratchDir}
}

// Promote is every WAL-G engine's promotion (walgkit.Promote), by walg-entrypoint
// promote. The member keeps its server id and log name, which are no standby's
// variables.
func (e *Engine) Promote(req *managedmodel.PromoteReq) (*managedmodel.PromoteSpec, error) {
	spec, err := walgkit.Promote(profile, req)
	return spec, hperrors.Wrap(err)
}
