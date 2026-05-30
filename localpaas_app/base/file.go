package base

type FileStatus string

const (
	FileStatusActive   FileStatus = "active"
	FileStatusPending  FileStatus = "pending"
	FileStatusDisabled FileStatus = "disabled"
)

var (
	AllFileStatuses = []FileStatus{FileStatusActive, FileStatusPending, FileStatusDisabled}
)

type FileType string

const (
	FileTypeSystemBackup FileType = "system-backup"
)

var (
	AllFileTypes = []FileType{FileTypeSystemBackup}
)

type FileStorageType string

const (
	FileStorageLocal FileStorageType = "local"
	FileStorageCloud FileStorageType = "cloud"
)

var (
	AllFileStorageTypes = []FileStorageType{FileStorageLocal, FileStorageCloud}
)
