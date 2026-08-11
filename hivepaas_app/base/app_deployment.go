package base

type BuildTool string

const (
	BuildToolDocker   BuildTool = "docker"
	BuildToolRailpack BuildTool = "railpack"
)

var (
	AllBuildTools = []BuildTool{BuildToolDocker, BuildToolRailpack}
)

type DeploymentMethod string

const (
	DeploymentMethodImage DeploymentMethod = `image`
	DeploymentMethodRepo  DeploymentMethod = "repo"
)

var (
	AllDeploymentMethods = []DeploymentMethod{DeploymentMethodImage, DeploymentMethodRepo}
)
