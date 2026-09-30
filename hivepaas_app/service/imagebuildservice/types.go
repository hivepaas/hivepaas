package imagebuildservice

import (
	"github.com/moby/moby/api/types/registry"
	"github.com/moby/moby/api/types/swarm"

	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/tasks/queue"
)

type ImageBuildReq struct {
	*queue.TaskExecData
	App *entity.App

	CommitHash string
	Dockerfile entity.DeploymentDockerfile
	// ImageTags are the tags this deployment asked for, without the environment
	// prefix. They are not stored anywhere: a release marker belongs to one build.
	ImageTags      []string
	PushToRegistry entity.ObjectID

	ImageBuildSettings *entity.ImageBuildSettings
	NoCache            bool
	BuildID            string

	CheckoutDir string
	TempDir     string // can be empty

	// Inputs are what the build reads from settings, already resolved. A build
	// run by an agent is always given them: the agent has no key to open a
	// stored secret with. Left nil, the build resolves them itself.
	Inputs *BuildInputs
}

// BuildInputs is everything a build takes from settings, with the secrets in it
// opened: the app resolves it, where the data encryption key is, and a build
// anywhere else works from it without reading a setting.
type BuildInputs struct {
	// EnvVars are the build's variables, passed as build arguments.
	EnvVars map[string]*string
	// RegistryAuths are the project's registries, by address: what the build
	// signs in to for the images it pulls.
	RegistryAuths map[string]registry.AuthConfig
	// PushRegistry is the registry the image is pushed to, nil when it is not.
	PushRegistry *registry.AuthConfig
	// Secrets are the secret values above, to keep out of the build's logs.
	Secrets []string
}

type ImageBuildResp struct {
	ImageTags []string
}

type BuildNodeResp struct {
	Node            *swarm.Node
	CurrentNodeID   string
	ReleaseNodeFunc func()
}

func (resp *BuildNodeResp) ReleaseNode() {
	if resp.ReleaseNodeFunc != nil {
		resp.ReleaseNodeFunc()
		resp.ReleaseNodeFunc = nil
	}
}
