package base

type DeploymentMethod string

const (
	DeploymentMethodImage DeploymentMethod = `image`
	DeploymentMethodRepo  DeploymentMethod = "repo"
	// DeploymentMethodFunction builds a function's code on its runtime's image,
	// with the Dockerfile HivePaaS writes for it.
	DeploymentMethodFunction DeploymentMethod = "function"
)

var (
	AllDeploymentMethods = []DeploymentMethod{DeploymentMethodImage, DeploymentMethodRepo, DeploymentMethodFunction}
)
