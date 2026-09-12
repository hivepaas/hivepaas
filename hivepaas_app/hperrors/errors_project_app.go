package hperrors

// Errors for projects
var (
	ErrProjectNotFound            = NewErr(ErrNotFound, "ERR_PROJECT_NOT_FOUND")
	ErrProjectEnvNotFound         = NewErr(ErrNotFound, "ERR_PROJECT_ENV_NOT_FOUND")
	ErrProjectInactive            = NewErr(ErrInactive, "ERR_PROJECT_INACTIVE")
	ErrProjectEnvInactive         = NewErr(ErrInactive, "ERR_PROJECT_ENV_INACTIVE")
	ErrProjectNameNotAllowed      = NewErr(ErrNotAllowed, "ERR_PROJECT_NAME_NOT_ALLOWED")
	ErrProjectNetworkUnavailable  = NewErr(ErrUnavailable, "ERR_PROJECT_NETWORK_UNAVAILABLE")
	ErrProjectEnvRemovalUnallowed = NewErr(ErrNotAllowed, "ERR_PROJECT_ENV_REMOVAL_UNALLOWED")
)

// Errors for apps
var (
	ErrAppNotFound                              = NewErr(ErrNotFound, "ERR_APP_NOT_FOUND")
	ErrAppInactive                              = NewErr(ErrInactive, "ERR_APP_INACTIVE")
	ErrAppIsCurrent                             = NewErr(ErrInactive, "ERR_APP_IS_CURRENT")
	ErrAppsNotInSameProjectEnv                  = NewErr(ErrNotAllowed, "ERR_APPS_NOT_IN_SAME_PROJECT_ENV")
	ErrAppCloneSettingsRequired                 = NewErr(ErrPreconditionRequired, "ERR_APP_CLONE_SETTING_REQUIRED")
	ErrMultiNodeClusterRequireRegistryForImages = NewErr(ErrPreconditionRequired, "ERR_MULTI_NODE_CLUSTER_REQUIRE_REGISTRY_FOR_IMAGES") //nolint:lll
	ErrDeploymentMethodRepoRequired             = NewErr(ErrUnconfigured, "ERR_DEPLOYMENT_METHOD_REPO_REQUIRED")
	ErrFeatureDisabled                          = NewErr(ErrInactive, "ERR_FEATURE_DISABLED")
)

// Errors for sources
var (
	ErrRepoNotFound             = NewErr(ErrNotFound, "ERR_REPO_NOT_FOUND")
	ErrRepoTypeUnsupported      = NewErr(ErrUnsupported, "ERR_REPO_TYPE_UNSUPPORTED")
	ErrRepoRefNotFound          = NewErr(ErrNotFound, "ERR_REPO_REF_NOT_FOUND")
	ErrPullRequestNotFound      = NewErr(ErrNotFound, "ERR_PULL_REQUEST_NOT_FOUND")
	ErrPullRequestInvalid       = NewErr(ErrValueInvalid, "ERR_PULL_REQUEST_INVALID")
	ErrGitTypeUnsupported       = NewErr(ErrUnsupported, "ERR_GIT_TYPE_UNSUPPORTED")
	ErrGitAuthMethodUnsupported = NewErr(ErrUnsupported, "ERR_GIT_AUTH_METHOD_UNSUPPORTED")
	ErrGitLogOutputUnexpected   = NewErr(ErrPreconditionFailed, "ERR_GIT_LOG_OUTPUT_UNEXPECTED")
)

// Errors for build
var ()
