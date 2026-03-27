package base

type FileKind string

const (
	FileKindSystemBackup  FileKind = "system-backup"
	FileKindSourceArchive FileKind = "source-archive"
)

var (
	AllFileKinds = []FileKind{FileKindSystemBackup, FileKindSourceArchive}
)

type FileStorageType string

const (
	FileStorageLocal FileStorageType = "local"
	FileStorageCloud FileStorageType = "cloud"
)

var (
	AllFileStorageTypes = []FileStorageType{FileStorageLocal, FileStorageCloud}
)
