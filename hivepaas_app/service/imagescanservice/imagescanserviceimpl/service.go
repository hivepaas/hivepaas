package imagescanserviceimpl

import (
	"context"
	"errors"

	"github.com/tiendc/gofn"

	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/infra/database"
	"github.com/hivepaas/hivepaas/hivepaas_app/infra/rediscache"
	agentimagescan "github.com/hivepaas/hivepaas/hivepaas_app/interface/agent/client/imagescanservice"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/logging"
	"github.com/hivepaas/hivepaas/hivepaas_app/repository"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/agentservice"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/imagescanservice"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/imagescanservice/imagescanrun"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/nodeexecservice"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/notificationservice"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/registryauthservice"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/systemappservice"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/volumeservice"
	"github.com/hivepaas/hivepaas/services/docker"
)

// defaultTrivyImage is what runs when the release names no Trivy image - a
// release file from before the field existed.
const defaultTrivyImage = "aquasec/trivy:0.75.0"

// scanner runs a scan on a node: in this process for its own, through the
// node's agent for another.
type scanner interface {
	Run(ctx context.Context, req *imagescanrun.RunReq) (*imagescanrun.RunResp, error)
}

type service struct {
	// db writes what a scan found on its own, outside the transaction of the
	// task that asked: a task holds one open for as long as it runs.
	db *database.DB

	appRepo       repository.AppRepo
	settingRepo   repository.SettingRepo
	imageScanRepo repository.ImageScanRepo
	taskRepo      repository.TaskRepo

	agentService        agentservice.Service
	nodeExecService     nodeexecservice.Service
	notificationService notificationservice.Service
	registryAuthService registryauthservice.Service
	volumeService       volumeservice.Service

	dockerManager docker.Manager
	redisLock     rediscache.Lock
	logger        logging.Logger

	localScanner   scanner
	newAgentClient func(agentAddr string) (agentimagescan.ImageScanServiceClient, error)
}

// New builds the image scan service. fx wires the arguments from the provider
// list in registry/provides.go.
//
//nolint:ireturn // the constructor of a service returns its interface
func New(
	db *database.DB,
	appRepo repository.AppRepo,
	settingRepo repository.SettingRepo,
	imageScanRepo repository.ImageScanRepo,
	taskRepo repository.TaskRepo,
	agentService agentservice.Service,
	nodeExecService nodeexecservice.Service,
	notificationService notificationservice.Service,
	registryAuthService registryauthservice.Service,
	volumeService volumeservice.Service,
	dockerManager docker.Manager,
	redisLock rediscache.Lock,
	logger logging.Logger,
) imagescanservice.Service {
	return &service{
		db:                  db,
		appRepo:             appRepo,
		settingRepo:         settingRepo,
		imageScanRepo:       imageScanRepo,
		taskRepo:            taskRepo,
		agentService:        agentService,
		nodeExecService:     nodeExecService,
		notificationService: notificationService,
		registryAuthService: registryAuthService,
		volumeService:       volumeService,
		dockerManager:       dockerManager,
		redisLock:           redisLock,
		logger:              logger,
		localScanner:        &imagescanrun.Runner{Docker: dockerManager},
		newAgentClient:      agentimagescan.NewImageScanServiceClient,
	}
}

func (s *service) Settings(ctx context.Context, db database.IDB) (
	*entity.Setting, *entity.ImageScanSettings, error) {
	setting, err := s.settingRepo.GetSingle(ctx, db, entity.NewObjectScopeGlobal(), base.SettingTypeImageScan, false)
	if err != nil {
		if errors.Is(err, hperrors.ErrNotFound) {
			return nil, nil, nil
		}
		return nil, nil, hperrors.Wrap(err)
	}
	cfg, err := setting.AsImageScanSettings()
	if err != nil {
		return nil, nil, hperrors.Wrap(err)
	}
	return setting, cfg, nil
}

// enabled is the configuration when scanning is on, and an error saying it is
// off otherwise.
func (s *service) enabled(ctx context.Context, db database.IDB) (*entity.ImageScanSettings, error) {
	_, cfg, err := s.Settings(ctx, db)
	if err != nil {
		return nil, err
	}
	if cfg == nil || !cfg.Enabled {
		return nil, hperrors.Wrap(hperrors.ErrImageScanOff)
	}
	return cfg, nil
}

// trivyImage is the Trivy this release runs.
func trivyImage() string {
	return gofn.Coalesce(systemappservice.CurrentRelease().TrivyImage, defaultTrivyImage)
}
