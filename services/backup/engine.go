package backup

import (
	"github.com/hivepaas/hivepaas/services/backup/backupmodel"
)

// Re-export constants
const (
	EngineTypeKopia = backupmodel.EngineTypeKopia
)

// Re-export variables and functions
var (
	AllEngineTypes         = backupmodel.AllEngineTypes
	DefaultCommandExecutor = backupmodel.DefaultCommandExecutor
)

// Re-export type aliases
type (
	EngineType           = backupmodel.EngineType
	Engine               = backupmodel.Engine
	Snapshot             = backupmodel.Snapshot
	InitRepoOptions      = backupmodel.InitRepoOptions
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
