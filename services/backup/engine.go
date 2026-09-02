package backup

import (
	"github.com/hivepaas/hivepaas/services/backup/backupmodel"
	"github.com/hivepaas/hivepaas/services/backup/kopia"
)

// Re-export constants
const (
	EngineTypeKopia = backupmodel.EngineTypeKopia
	CompressionNone = backupmodel.CompressionNone
)

// Re-export variables and functions
var (
	AllEngineTypes         = backupmodel.AllEngineTypes
	NewRepoOptions         = backupmodel.NewRepoOptions
	DefaultCommandExecutor = backupmodel.DefaultCommandExecutor

	// AllCompressionAlgorithms are the compression algorithms a repository accepts. Anything else is
	// rejected by the engine, so it is worth catching before the request reaches it.
	AllCompressionAlgorithms = map[EngineType][]string{
		EngineTypeKopia: kopia.AllCompressionAlgorithms,
	}
)

// Re-export type aliases
type (
	EngineType           = backupmodel.EngineType
	Engine               = backupmodel.Engine
	Snapshot             = backupmodel.Snapshot
	InitRepoOptions      = backupmodel.InitRepoOptions
	RepoOptions          = backupmodel.RepoOptions
	RepoConfig           = backupmodel.RepoConfig
	BackupOptions        = backupmodel.BackupOptions
	ListSnapshotsOptions = backupmodel.ListSnapshotsOptions
	ListSnapshotsResult  = backupmodel.ListSnapshotsResult
	RetentionPolicy      = backupmodel.RetentionPolicy
	Storage              = backupmodel.Storage
	StorageS3            = backupmodel.StorageS3
	StorageLocal         = backupmodel.StorageLocal
	CommandExecReq       = backupmodel.CommandExecReq
	CommandExecResp      = backupmodel.CommandExecResp
	CommandExecutor      = backupmodel.CommandExecutor
)
