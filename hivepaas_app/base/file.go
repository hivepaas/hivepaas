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
	FileTypeRepoCache      FileType = "repo-cache"
	FileTypeBuildSource    FileType = "build-source"
	FileTypeSchedJobOutput FileType = "sched-job-output"
)

var (
	AllFileTypes = []FileType{FileTypeDataFile, FileTypeSystemBackup, FileTypeRepoCache, FileTypeBuildSource,
		FileTypeSchedJobOutput}
)

type FileKind string

const (
	FileKindBackupClickhouse FileKind = "clickhouse-backup"
	FileKindBackupOracle     FileKind = "oracle-backup"
	FileKindBackupMaria      FileKind = "maria-backup"
	FileKindBackupMongo      FileKind = "mongo-backup"
	FileKindBackupMysql      FileKind = "mysql-backup"
	FileKindBackupPostgres   FileKind = "postgres-backup"
	FileKindBackupRedis      FileKind = "redis-backup"
	FileKindBackupSqlServer  FileKind = "sql-server-backup"
)

var (
	AllFileKinds = []FileKind{FileKindBackupClickhouse, FileKindBackupOracle, FileKindBackupMaria, FileKindBackupMongo,
		FileKindBackupMysql, FileKindBackupPostgres, FileKindBackupRedis, FileKindBackupSqlServer}
)

type FileStorageType string

const (
	FileStorageLocal FileStorageType = "local"
	FileStorageCloud FileStorageType = "cloud"
)

var (
	AllFileStorageTypes = []FileStorageType{FileStorageLocal, FileStorageCloud}
)
