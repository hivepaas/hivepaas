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
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/logging"
	"github.com/hivepaas/hivepaas/hivepaas_app/repository"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/containerexecservice"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/imagescanservice"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/nodeexecservice"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/notificationservice"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/systemappservice"
	"github.com/hivepaas/hivepaas/services/docker"
)

type service struct {
	// db writes what a scan found on its own, outside the transaction of the
	// task that asked: a task holds one open for as long as it runs.
	db *database.DB

	appRepo       repository.AppRepo
	settingRepo   repository.SettingRepo
	imageScanRepo repository.ImageScanRepo
	taskRepo      repository.TaskRepo

	systemAppService     systemappservice.Service
	containerExecService containerexecservice.Service
	nodeExecService      nodeexecservice.Service
	notificationService  notificationservice.Service

	dockerManager docker.Manager
	redisLock     rediscache.Lock
	logger        logging.Logger
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
	systemAppService systemappservice.Service,
	containerExecService containerexecservice.Service,
	nodeExecService nodeexecservice.Service,
	notificationService notificationservice.Service,
	dockerManager docker.Manager,
	redisLock rediscache.Lock,
	logger logging.Logger,
) imagescanservice.Service {
	return &service{
		db:                   db,
		appRepo:              appRepo,
		settingRepo:          settingRepo,
		imageScanRepo:        imageScanRepo,
		taskRepo:             taskRepo,
		systemAppService:     systemAppService,
		containerExecService: containerExecService,
		nodeExecService:      nodeExecService,
		notificationService:  notificationService,
		dockerManager:        dockerManager,
		redisLock:            redisLock,
		logger:               logger,
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

// trivyImage is the Trivy this release runs: the server and the runner, which
// must be one version.
func trivyImage() string {
	return gofn.Coalesce(systemappservice.CurrentRelease().TrivyImage, defaultTrivyImage)
}
