package settingsrevertserviceimpl

import (
	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/logging"
	"github.com/hivepaas/hivepaas/hivepaas_app/repository"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/approutingservice"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/hpappservice"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/settingsrevertservice"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/traefikservice"
	"github.com/hivepaas/hivepaas/services/docker"
)

func New(
	settingRepo repository.SettingRepo,

	dockerManager docker.Manager,

	appRoutingService approutingservice.Service,
	hpAppService hpappservice.Service,
	traefikService traefikservice.Service,
	logger logging.Logger,
) settingsrevertservice.Service {
	s := &service{
		settingRepo:       settingRepo,
		dockerManager:     dockerManager,
		appRoutingService: appRoutingService,
		hpAppService:      hpAppService,
		traefikService:    traefikService,
		logger:            logger,
	}

	// One entry per setting type that can be put on trial. A type missing from
	// here is refused rather than guessed at: a revert that applies the wrong
	// kind of change is worse than one that does not happen, because the operator
	// is still expecting the one they asked for.
	s.reverters = map[base.SettingType]reverter{
		base.SettingTypeAppRouting:      s.revertAppRouting,
		base.SettingTypeHivePaaSService: s.revertHivePaaSService,
	}

	return s
}

type service struct {
	settingRepo repository.SettingRepo

	dockerManager docker.Manager

	appRoutingService approutingservice.Service
	hpAppService      hpappservice.Service
	traefikService    traefikservice.Service
	logger            logging.Logger

	reverters map[base.SettingType]reverter
}
