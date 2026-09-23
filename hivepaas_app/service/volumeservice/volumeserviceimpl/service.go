package volumeserviceimpl

import (
	"context"

	"github.com/moby/moby/api/types/mount"

	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/logging"
	"github.com/hivepaas/hivepaas/hivepaas_app/repository"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/agentservice"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/hpappservice"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/volumeservice"
	"github.com/hivepaas/hivepaas/services/docker"
)

func New(
	dockerManager docker.Manager,
	hpAppService hpappservice.Service,
	agentService agentservice.Service,

	settingRepo repository.SettingRepo,

	logger logging.Logger,
) volumeservice.Service {
	svc := &service{
		dockerManager: dockerManager,
		hpAppService:  hpAppService,
		agentService:  agentService,

		settingRepo: settingRepo,
		logger:      logger,
	}
	svc.makeSubDirInHost = svc.MakeSubDirInHost
	svc.ensureVolumePermissions = svc.EnsureVolumePermissions
	return svc
}

type service struct {
	dockerManager docker.Manager
	hpAppService  hpappservice.Service
	agentService  agentservice.Service

	settingRepo repository.SettingRepo

	logger logging.Logger

	// makeSubDirInHost and ensureVolumePermissions reach the host through a
	// throwaway container. They are fields so a test of mount building can stand
	// in for the host without a docker daemon.
	makeSubDirInHost        func(ctx context.Context, baseDirInHost, subpath string, requireBaseDirExist bool) error
	ensureVolumePermissions func(ctx context.Context, volMount *mount.Mount, subpaths ...string) error
}
