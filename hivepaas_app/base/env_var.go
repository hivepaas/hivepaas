package base

import "maps"

const (
	SecretRefInEnvMaxSize = 10 * 1024 // 10 KB
)

type EnvVarKind string

const (
	EnvVarKindRuntime EnvVarKind = "runtime"
	EnvVarKindShared  EnvVarKind = "shared"
	EnvVarKindBuild   EnvVarKind = "build"
)

const (
	// Project system envs
	ProjectSystemEnvVarName = "HIVEPAAS_PROJECT_NAME"
	ProjectSystemEnvVarID   = "HIVEPAAS_PROJECT_ID"
)

var (
	mapProjectUnallowedVar = map[string]struct{}{
		ProjectSystemEnvVarName: {},
		ProjectSystemEnvVarID:   {},
	}

	mapProjectSecretVar = map[string]struct{}{}
)

func IsProjectRuntimeEnvAllowed(env string) bool {
	_, exists := mapProjectUnallowedVar[env]
	return !exists
}

func IsProjectBuildEnvAllowed(env string) bool {
	_, exists := mapProjectUnallowedVar[env]
	return !exists
}

func IsProjectSecretEnv(env string) bool {
	_, exists := mapProjectSecretVar[env]
	return exists
}

const (
	// App system envs
	AppSystemEnvVarHost   = "HIVEPAAS_HOST"
	AppSystemEnvVarPort   = "HIVEPAAS_PORT"
	AppSystemEnvVarDomain = "HIVEPAAS_DOMAIN"
	// AppSystemEnvVarAppURL is where the app answers from outside - scheme and
	// domain - and is empty while it has no domain. An app that has to be told
	// its own address usually wants this rather than the domain alone, and the
	// empty value is what such a setting expects when there is nothing to tell
	// it: many of them refuse to start on a scheme with no host after it.
	AppSystemEnvVarAppURL = "HIVEPAAS_APP_URL"
	AppSystemEnvVarEnv    = "HIVEPAAS_ENV"
	AppSystemEnvVarName   = "HIVEPAAS_APP_NAME"
	AppSystemEnvVarID     = "HIVEPAAS_APP_ID"

	AppSystemEnvVarUser         = "HIVEPAAS_USER"
	AppSystemEnvVarPassword     = "HIVEPAAS_PASSWORD"      //nolint:gosec // G101: env name
	AppSystemEnvVarRootPassword = "HIVEPAAS_ROOT_PASSWORD" //nolint:gosec // G101: env name
	AppSystemEnvVarDatabaseName = "HIVEPAAS_DATABASE_NAME"
	AppSystemEnvVarSSLMode      = "HIVEPAAS_SSL_MODE"

	AppSystemEnvVarKeyID  = "HIVEPAAS_KEY_ID"
	AppSystemEnvVarSecret = "HIVEPAAS_SECRET" //nolint:gosec // G101: env name
	AppSystemEnvVarBucket = "HIVEPAAS_BUCKET"
	AppSystemEnvVarRegion = "HIVEPAAS_REGION"

	// How a cache is tuned, for the app's own container to read. These are not
	// credentials and no other app has any use for them, so unlike the password
	// they are not shared - they exist so that the cache settings on the App Kind
	// screen reach the server instead of only being recorded.
	AppSystemEnvVarMaxMemory       = "HIVEPAAS_MAX_MEMORY"
	AppSystemEnvVarEvictionRule    = "HIVEPAAS_EVICTION_RULE"
	AppSystemEnvVarPersistenceMode = "HIVEPAAS_PERSISTENCE_MODE"
)

// AppCommonSharedEnvVars are shared by every app, whatever its kind.
var AppCommonSharedEnvVars = []string{
	AppSystemEnvVarHost, AppSystemEnvVarPort, AppSystemEnvVarDomain, AppSystemEnvVarAppURL,
	AppSystemEnvVarEnv, AppSystemEnvVarName, AppSystemEnvVarID,
}

// AppKindSharedEnvVars are what an app of a kind shares on top of the common
// variables - what another app's ${<app>.VAR} may name. envvarserviceimpl
// publishes exactly these, and a test there holds the two together.
func AppKindSharedEnvVars(category AppCategory) []string {
	switch category {
	case AppCategoryDatabase:
		return []string{AppSystemEnvVarUser, AppSystemEnvVarPassword, AppSystemEnvVarDatabaseName,
			AppSystemEnvVarSSLMode}
	case AppCategoryCache:
		return []string{AppSystemEnvVarPassword}
	case AppCategoryStorage:
		return []string{AppSystemEnvVarKeyID, AppSystemEnvVarSecret, AppSystemEnvVarBucket, AppSystemEnvVarRegion}
	case AppCategoryWebapp:
		return nil
	}
	return nil
}

var (
	mapAppUnallowedVar = func() map[string]struct{} {
		theMap := map[string]struct{}{
			AppSystemEnvVarHost:         {},
			AppSystemEnvVarPort:         {},
			AppSystemEnvVarDomain:       {},
			AppSystemEnvVarAppURL:       {},
			AppSystemEnvVarEnv:          {},
			AppSystemEnvVarName:         {},
			AppSystemEnvVarID:           {},
			AppSystemEnvVarUser:         {},
			AppSystemEnvVarPassword:     {},
			AppSystemEnvVarRootPassword: {},
			AppSystemEnvVarDatabaseName: {},
			AppSystemEnvVarSSLMode:      {},
			AppSystemEnvVarKeyID:        {},
			AppSystemEnvVarSecret:       {},
			AppSystemEnvVarBucket:       {},
			AppSystemEnvVarRegion:       {},

			AppSystemEnvVarMaxMemory:       {},
			AppSystemEnvVarEvictionRule:    {},
			AppSystemEnvVarPersistenceMode: {},
		}
		maps.Copy(theMap, mapProjectUnallowedVar)
		return theMap
	}()

	mapAppSecretVar = map[string]struct{}{
		AppSystemEnvVarPassword:     {},
		AppSystemEnvVarRootPassword: {},
		AppSystemEnvVarSecret:       {},
	}
)

func IsAppRuntimeEnvAllowed(env string) bool {
	_, exists := mapAppUnallowedVar[env]
	return !exists
}

func IsAppSharedEnvAllowed(env string) bool {
	_, exists := mapAppUnallowedVar[env]
	return !exists
}

func IsAppSharedEnvSettable(env string) bool {
	_, exists := mapAppUnallowedVar[env]
	return !exists
}

func IsAppBuildEnvAllowed(env string) bool {
	_, exists := mapAppUnallowedVar[env]
	return !exists
}

func IsAppSecretEnv(env string) bool {
	_, exists := mapAppSecretVar[env]
	return exists
}
