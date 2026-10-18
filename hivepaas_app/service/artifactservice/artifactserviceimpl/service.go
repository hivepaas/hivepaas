package artifactserviceimpl

import (
	"github.com/hivepaas/hivepaas/hivepaas_app/config"
	agentartifactclient "github.com/hivepaas/hivepaas/hivepaas_app/interface/agent/client/artifactservice"
	"github.com/hivepaas/hivepaas/hivepaas_app/repository"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/agentservice"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/artifactservice"
	"github.com/hivepaas/hivepaas/services/docker"
)

type service struct {
	dockerManager docker.Manager
	agentService  agentservice.Service
	settingRepo   repository.SettingRepo
	appRepo       repository.AppRepo
	newClient     func(agentAddr string) (agentartifactclient.ArtifactServiceClient, error)
	// mirror is the base URL of a mirror holding artifacts by their hash, read
	// when it is needed so a changed config applies without a restart.
	mirror func() string
}

func New(
	dockerManager docker.Manager,
	agentService agentservice.Service,
	settingRepo repository.SettingRepo,
	appRepo repository.AppRepo,
) artifactservice.Service {
	return &service{
		dockerManager: dockerManager,
		agentService:  agentService,
		settingRepo:   settingRepo,
		appRepo:       appRepo,
		newClient:     agentartifactclient.NewArtifactServiceClient,
		mirror:        func() string { return config.Current().Artifacts.Mirror },
	}
}
