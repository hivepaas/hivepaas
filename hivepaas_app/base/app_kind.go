package base

type AppCategory string

const (
	AppCategoryDatabase AppCategory = "database"
	AppCategoryWebapp   AppCategory = "webapp"
	AppCategoryCache    AppCategory = "cache"
	AppCategoryStorage  AppCategory = "storage"
	// AppCategoryFunction is a function: a handler HivePaaS runs on a runtime of
	// its own. An app is one from its creation, and stays one.
	AppCategoryFunction AppCategory = "function"
)

var (
	AllAppCategories = []AppCategory{AppCategoryDatabase, AppCategoryWebapp, AppCategoryCache,
		AppCategoryStorage, AppCategoryFunction}
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
