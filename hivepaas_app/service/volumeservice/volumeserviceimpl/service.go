package volumeserviceimpl

import (
	"context"

	"github.com/moby/moby/api/types/mount"

	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/logging"
	"github.com/hivepaas/hivepaas/hivepaas_app/repository"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/agentservice"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/hpappservice"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/nodeexecservice"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/volumeservice"
	"github.com/hivepaas/hivepaas/services/docker"
)

func New(
	dockerManager docker.Manager,
	hpAppService hpappservice.Service,
	agentService agentservice.Service,
	nodeExecService nodeexecservice.Service,

	settingRepo repository.SettingRepo,

	logger logging.Logger,
) volumeservice.Service {
	svc := &service{
		dockerManager:   dockerManager,
		hpAppService:    hpAppService,
		agentService:    agentService,
		nodeExecService: nodeExecService,

		settingRepo: settingRepo,
		logger:      logger,
	}
	svc.makeSubDirInHost = svc.makeSubDirOnNode
	svc.ensureVolumePermissions = svc.ensureVolumePermissionsOnNode
	return svc
}

type service struct {
	dockerManager docker.Manager
	hpAppService  hpappservice.Service
	agentService  agentservice.Service
	// nodeExecService reaches the node a volume's data is on, for storage this
	// node cannot open itself.
	nodeExecService nodeexecservice.Service

	settingRepo repository.SettingRepo

	logger logging.Logger

	// makeSubDirInHost and ensureVolumePermissions reach the host of the volume's
	// node - vol's pin; this node's for a volume with none - through a throwaway
	// container or that node's agent. They are fields so a test of mount building
	// can stand in for the hosts without a docker daemon.
	makeSubDirInHost func(ctx context.Context, vol *entity.ClusterVolume, baseDirInHost, subpath string,
		requireBaseDirExist bool) error
	ensureVolumePermissions func(ctx context.Context, vol *entity.ClusterVolume, volMount *mount.Mount,
		subpaths ...string) error
}
