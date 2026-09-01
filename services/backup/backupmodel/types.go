package backupmodel

import (
	"time"
)

// EngineType represents the type of backup engine.
type EngineType string

const (
	EngineTypeKopia EngineType = "kopia"
)

// AllEngineTypes contains all supported engine types.
var AllEngineTypes = []EngineType{EngineTypeKopia}

// Snapshot represents a standardized backup snapshot across engines.
type Snapshot struct {
	ID        string    `json:"id"`
	ShortID   string    `json:"shortId"`
	Time      time.Time `json:"time"`
	Tags      []string  `json:"tags"`
	Paths     []string  `json:"paths"`
	Hostname  string    `json:"hostname"`
	SizeBytes int64     `json:"sizeBytes,omitempty"`
}

// InitRepoOptions contains the settings applied while creating a new repository.
type InitRepoOptions struct {
	Description string `json:"description,omitempty"`
	PackSizeMB  int    `json:"packSizeMb,omitempty"`
	Compression string `json:"compression,omitempty"`
}

// BackupOptions contains parameters for running a backup operation.
type BackupOptions struct {
	Tags     []string `json:"tags,omitempty"`
	Hostname string   `json:"hostname,omitempty"`
	PackSize int      `json:"packSize,omitempty"` // in MB
}

type BackupResult struct {
	Item *Snapshot `json:"item,omitempty"`
}

type RestoreOptions struct {
}

type RestoreResult struct {
}

// ListSnapshotsOptions contains filtering and pagination options for listing snapshots.
type ListSnapshotsOptions struct {
	Tags     []string   `json:"tags,omitempty"`
	Path     string     `json:"path,omitempty"`
	Hostname string     `json:"hostname,omitempty"`
	Since    *time.Time `json:"since,omitempty"`
	Until    *time.Time `json:"until,omitempty"`
	Limit    int        `json:"limit,omitempty"`
}

type ListSnapshotsResult struct {
	Items []*Snapshot `json:"items,omitempty"`
}

type GetSnapshotResult struct {
	Item *Snapshot `json:"item,omitempty"`
}

type DeleteSnapshotResult struct {
}

type PruneResult struct {
}

// RetentionPolicy defines the snapshot retention and pruning rules.
type RetentionPolicy struct {
	KeepLast    int `json:"keepLast,omitempty"`
	KeepDaily   int `json:"keepDaily,omitempty"`
	KeepWeekly  int `json:"keepWeekly,omitempty"`
	KeepMonthly int `json:"keepMonthly,omitempty"`
}

type Storage struct {
	RepositoryPassword string        `json:"repositoryPassword"`
	StorageS3          *StorageS3    `json:"storageS3,omitempty"`
	StorageLocal       *StorageLocal `json:"storageLocal,omitempty"`

	// ConfigFile is the engine config file holding the repository connection state.
	// Every repository must use its own file, otherwise operations on different repositories
	// overwrite each other's connection state as they all share the engine default config file.
	ConfigFile string `json:"configFile,omitempty"`
}

type StorageS3 struct {
	Endpoint       string `json:"endpoint,omitempty"`
	Region         string `json:"region"`
	Bucket         string `json:"bucket"`
	Prefix         string `json:"prefix,omitempty"`
	AccessKey      string `json:"accessKey"`
	SecretKey      string `json:"secretKey"`
	ForcePathStyle bool   `json:"forcePathStyle"`
}

type StorageLocal struct {
	Path      string `json:"path"`
	NodeID    string `json:"nodeId,omitempty"`
	NodeLabel string `json:"nodeLabel,omitempty"`
}
