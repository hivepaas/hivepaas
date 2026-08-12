package base

type DeploymentMethod string

const (
	DeploymentMethodImage DeploymentMethod = `image`
	DeploymentMethodRepo  DeploymentMethod = "repo"
)

var (
	AllDeploymentMethods = []DeploymentMethod{DeploymentMethodImage, DeploymentMethodRepo}
)
