package base

type BuildTool string

const (
	BuildToolDocker   BuildTool = "docker"
	BuildToolNixpacks BuildTool = "nixpacks"
)

var (
	AllBuildTools = []BuildTool{BuildToolDocker, BuildToolNixpacks}
)

type DeploymentMethod string

const (
	DeploymentMethodImage DeploymentMethod = `image`
	DeploymentMethodRepo  DeploymentMethod = "repo"
)

var (
	AllDeploymentMethods = []DeploymentMethod{DeploymentMethodImage, DeploymentMethodRepo}
)
