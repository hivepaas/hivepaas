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
)

func IsProjectRuntimeEnvAllowed(env string) bool {
	_, exists := mapProjectUnallowedVar[env]
	return !exists
}

func IsProjectBuildEnvAllowed(env string) bool {
	_, exists := mapProjectUnallowedVar[env]
	return !exists
}

const (
	// App system envs
	AppSystemEnvVarHost   = "HIVEPAAS_HOST"
	AppSystemEnvVarPort   = "HIVEPAAS_PORT"
	AppSystemEnvVarDomain = "HIVEPAAS_DOMAIN"
	AppSystemEnvVarEnv    = "HIVEPAAS_ENV"
	AppSystemEnvVarName   = "HIVEPAAS_APP_NAME"
	AppSystemEnvVarID     = "HIVEPAAS_APP_ID"
)

var (
	mapAppUnallowedVar = func() map[string]struct{} {
		theMap := map[string]struct{}{
			AppSystemEnvVarHost:   {},
			AppSystemEnvVarPort:   {},
			AppSystemEnvVarDomain: {},
			AppSystemEnvVarEnv:    {},
			AppSystemEnvVarName:   {},
			AppSystemEnvVarID:     {},
		}
		maps.Copy(theMap, mapProjectUnallowedVar)
		return theMap
	}()
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
