package base

const (
	// 0755 grants read/write/execute for owner, read/execute for group/others
	DirModeDefault = 0755
)

type FileStatus string

const (
	FileStatusActive   FileStatus = "active"
	FileStatusPending  FileStatus = "pending"
	FileStatusDisabled FileStatus = "disabled"
	FileStatusDeleting FileStatus = "deleting"
)

var (
	AllFileStatuses = []FileStatus{FileStatusActive, FileStatusPending, FileStatusDisabled, FileStatusDeleting}
)

type FileType string

const (
	FileTypeDataFile       FileType = "data-file"
	FileTypeSystem         FileType = "system"
	FileTypeCache          FileType = "cache"
	FileTypeTmp            FileType = "tmp"
	FileTypeSchedJobOutput FileType = "sched-job-output"
)

var (
	AllFileTypes = []FileType{FileTypeDataFile, FileTypeSystem, FileTypeCache, FileTypeTmp,
		FileTypeSchedJobOutput}
)

type FileKind string

const (
	// File kinds of type `data-file`
	FileKindBackupClickhouse FileKind = "clickhouse-backup"
	FileKindBackupOracle     FileKind = "oracle-backup"
	FileKindBackupMaria      FileKind = "maria-backup"
	FileKindBackupMongo      FileKind = "mongo-backup"
	FileKindBackupMysql      FileKind = "mysql-backup"
	FileKindBackupPostgres   FileKind = "postgres-backup"
	FileKindBackupRedis      FileKind = "redis-backup"
	FileKindBackupSqlServer  FileKind = "sql-server-backup"

	// File kinds of type `cache`, `tmp`
	FileKindSourceCode FileKind = "source-code"

	// File kinds of type `system`
	FileKindSystemBackup FileKind = "system-backup"
)

var (
	AllFileDataKinds = []FileKind{FileKindBackupClickhouse, FileKindBackupOracle, FileKindBackupMaria, FileKindBackupMongo,
		FileKindBackupMysql, FileKindBackupPostgres, FileKindBackupRedis, FileKindBackupSqlServer}

	AllFileCacheKinds = []FileKind{FileKindSourceCode}

	AllFileSystemBackupKinds = []FileKind{FileKindSystemBackup}
)

type FileStorageType string

const (
	FileStorageLocal FileStorageType = "local"
	FileStorageCloud FileStorageType = "cloud"
)

var (
	AllFileStorageTypes = []FileStorageType{FileStorageLocal, FileStorageCloud}
)
