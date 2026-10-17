package artifactserviceimpl

import (
	"github.com/hivepaas/hivepaas/hivepaas_app/config"
	agentartifactclient "github.com/hivepaas/hivepaas/hivepaas_app/interface/agent/client/artifactservice"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/agentservice"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/artifactservice"
	"github.com/hivepaas/hivepaas/services/docker"
)

type service struct {
	dockerManager docker.Manager
	agentService  agentservice.Service
	newClient     func(agentAddr string) (agentartifactclient.ArtifactServiceClient, error)
	// mirror is the base URL of a mirror holding artifacts by their hash, read
	// when it is needed so a changed config applies without a restart.
	mirror func() string
}

func New(
	dockerManager docker.Manager,
	agentService agentservice.Service,
) artifactservice.Service {
	return &service{
		dockerManager: dockerManager,
		agentService:  agentService,
		newClient:     agentartifactclient.NewArtifactServiceClient,
		mirror:        func() string { return config.Current().Artifacts.Mirror },
	}
}
