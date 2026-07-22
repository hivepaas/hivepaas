package base

type EnvVarKind string

const (
	EnvVarKindRuntime EnvVarKind = "runtime"
	EnvVarKindShared  EnvVarKind = "shared"
	EnvVarKindBuild   EnvVarKind = "build"
)

const (
	// App system envs
	AppSystemEnvVarName = "HIVEPAAS_APP_NAME"
	AppSystemEnvVarID   = "HIVEPAAS_APP_ID"
	AppSystemEnvVarEnv  = "HIVEPAAS_APP_ENV"
	AppSystemEnvVarHost = "HIVEPAAS_HOST"
	AppSystemEnvVarPort = "HIVEPAAS_PORT"
)

func IsAppRuntimeEnvAllowed(env string) bool {
	return true
}

func IsAppBuildEnvAllowed(env string) bool {
	return true
}

const (
	// Project system envs
	ProjectSystemEnvVarName = "HIVEPAAS_PROJECT_NAME"
	ProjectSystemEnvVarID   = "HIVEPAAS_PROJECT_ID"
)

func IsProjectRuntimeEnvAllowed(env string) bool {
	return true
}

func IsProjectBuildEnvAllowed(env string) bool {
	return true
}
