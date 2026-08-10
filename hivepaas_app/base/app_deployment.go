package base

type BuildTool string

const (
	BuildToolDocker   BuildTool = "docker"
	BuildToolRailPack BuildTool = "railpack"
)

var (
	AllBuildTools = []BuildTool{BuildToolDocker, BuildToolRailPack}
)

type DeploymentMethod string

const (
	DeploymentMethodImage DeploymentMethod = `image`
	DeploymentMethodRepo  DeploymentMethod = "repo"
)

var (
	AllDeploymentMethods = []DeploymentMethod{DeploymentMethodImage, DeploymentMethodRepo}
)
