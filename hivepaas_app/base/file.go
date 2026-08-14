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
	FileTypeSystemBackup   FileType = "system-backup"
	FileTypeCache          FileType = "cache"
	FileTypeBuildSource    FileType = "build-source"
	FileTypeSchedJobOutput FileType = "sched-job-output"
)

var (
	AllFileTypes = []FileType{FileTypeDataFile, FileTypeSystemBackup, FileTypeCache, FileTypeBuildSource,
		FileTypeSchedJobOutput}
)

type FileKind string

const (
	// File kinds of type data-file
	FileKindBackupClickhouse FileKind = "clickhouse-backup"
	FileKindBackupOracle     FileKind = "oracle-backup"
	FileKindBackupMaria      FileKind = "maria-backup"
	FileKindBackupMongo      FileKind = "mongo-backup"
	FileKindBackupMysql      FileKind = "mysql-backup"
	FileKindBackupPostgres   FileKind = "postgres-backup"
	FileKindBackupRedis      FileKind = "redis-backup"
	FileKindBackupSqlServer  FileKind = "sql-server-backup"

	// File kinds of type cache
	FileKindSourceCode FileKind = "source-code"
)

var (
	AllFileDataKinds = []FileKind{FileKindBackupClickhouse, FileKindBackupOracle, FileKindBackupMaria, FileKindBackupMongo,
		FileKindBackupMysql, FileKindBackupPostgres, FileKindBackupRedis, FileKindBackupSqlServer}

	AllFileCacheKinds = []FileKind{FileKindSourceCode}
)

type FileStorageType string

const (
	FileStorageLocal FileStorageType = "local"
	FileStorageCloud FileStorageType = "cloud"
)

var (
	AllFileStorageTypes = []FileStorageType{FileStorageLocal, FileStorageCloud}
)
