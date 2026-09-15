package base

type AppCategory string

const (
	AppCategoryDatabase AppCategory = "database"
	AppCategoryWebapp   AppCategory = "webapp"
	AppCategoryCache    AppCategory = "cache"
)

var (
	AllAppCategories = []AppCategory{AppCategoryDatabase, AppCategoryWebapp, AppCategoryCache}
)

type DatabaseSSLMode string

const (
	DatabaseSslModeDisable    DatabaseSSLMode = "disable"
	DatabaseSslModePrefer     DatabaseSSLMode = "prefer"
	DatabaseSslModeRequire    DatabaseSSLMode = "require"
	DatabaseSslModeVerifyCA   DatabaseSSLMode = "verify-ca"
	DatabaseSslModeVerifyFull DatabaseSSLMode = "verify-full"
)

var (
	AllDatabaseSslModes = []DatabaseSSLMode{DatabaseSslModeDisable, DatabaseSslModePrefer,
		DatabaseSslModeRequire, DatabaseSslModeVerifyCA, DatabaseSslModeVerifyFull}
)
