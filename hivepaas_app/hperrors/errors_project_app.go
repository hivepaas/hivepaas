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
	// The project HivePaaS runs in, and what in it runs this installation: not
	// to be deleted, disabled or stopped by hand.
	ErrSystemProjectProtected    = NewErr(ErrNotAllowed, "ERR_SYSTEM_PROJECT_PROTECTED")
	ErrSystemProjectEnvProtected = NewErr(ErrNotAllowed, "ERR_SYSTEM_PROJECT_ENV_PROTECTED")
	ErrSystemAppProtected        = NewErr(ErrNotAllowed, "ERR_SYSTEM_APP_PROTECTED")
)

// Errors for apps
var (
	ErrAppNotFound                              = NewErr(ErrNotFound, "ERR_APP_NOT_FOUND")
	ErrAppInactive                              = NewErr(ErrInactive, "ERR_APP_INACTIVE")
	ErrAppIsCurrent                             = NewErr(ErrInactive, "ERR_APP_IS_CURRENT")
	ErrAppsNotInSameProjectEnv                  = NewErr(ErrNotAllowed, "ERR_APPS_NOT_IN_SAME_PROJECT_ENV")
	ErrAppServiceUnavailable                    = NewErr(ErrUnavailable, "ERR_APP_SERVICE_UNAVAILABLE")
	ErrAppCloneSettingsRequired                 = NewErr(ErrPreconditionRequired, "ERR_APP_CLONE_SETTING_REQUIRED")
	ErrMultiNodeClusterRequireRegistryForImages = NewErr(ErrPreconditionRequired, "ERR_MULTI_NODE_CLUSTER_REQUIRE_REGISTRY_FOR_IMAGES") //nolint:lll
	ErrDeploymentMethodRepoRequired             = NewErr(ErrUnconfigured, "ERR_DEPLOYMENT_METHOD_REPO_REQUIRED")
	ErrFeatureDisabled                          = NewErr(ErrInactive, "ERR_FEATURE_DISABLED")
	// A build variable that uses a secret, which the Dockerfile declares with ARG:
	// a build argument's value is written into the image's history.
	ErrBuildSecretDeclaredAsArg = NewErr(ErrNotAllowed, "ERR_BUILD_SECRET_DECLARED_AS_ARG")
	// A function whose runtime the running release names no image for.
	ErrFunctionRuntimeUnavailable = NewErr(ErrUnsupported, "ERR_FUNCTION_RUNTIME_UNAVAILABLE")
	// A function is deployed from its code, and only a function is: the
	// deployment method follows the app's kind.
	ErrDeploymentMethodFunctionRequired  = NewErr(ErrPreconditionFailed, "ERR_DEPLOYMENT_METHOD_FUNCTION_REQUIRED")
	ErrDeploymentMethodFunctionUnallowed = NewErr(ErrPreconditionFailed, "ERR_DEPLOYMENT_METHOD_FUNCTION_UNALLOWED")
	// An app is a function from its creation, and stays one.
	ErrAppKindFunctionUnchangeable = NewErr(ErrNonEditable, "ERR_APP_KIND_FUNCTION_UNCHANGEABLE")
	// What only a function has: a test run, say.
	ErrAppNotFunction = NewErr(ErrPreconditionFailed, "ERR_APP_NOT_FUNCTION")
)

// Errors for autoscale
var (
	// What the app scales on cannot be read now: it is not turned on.
	ErrAutoscaleUnreadable = NewErr(ErrPreconditionFailed, "ERR_AUTOSCALE_UNREADABLE")
	// An app other than a function scales on its requests, its CPU or both.
	ErrAutoscaleNoSignal = NewErr(ErrPreconditionFailed, "ERR_AUTOSCALE_NO_SIGNAL")
	// An app publishing a port in host mode runs one replica a node at most.
	ErrAutoscaleHostPorts = NewErr(ErrPreconditionFailed, "ERR_AUTOSCALE_HOST_PORTS")
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
