package specserviceimpl

import (
	"context"

	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/infra/database"
	"github.com/hivepaas/hivepaas/hivepaas_app/repository"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/appdeploymentservice"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/appprovisionservice"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/approutingservice"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/appservice"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/clustersecretservice"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/clusterservice"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/domainservice"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/envvarservice"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/networkservice"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/projectservice"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/specservice"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/specservice/specmodel"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/sslservice"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/volumeservice"
	"github.com/hivepaas/hivepaas/hivepaas_app/tasks/queue"
)

// New builds the spec service.
//
// Export resolves a Docker network id to its name from the cluster-network
// settings alone, which sync writes with RefID set to the Docker id and Name to
// the Docker name - no round trip to Docker. Import asks networkservice only
// what an env's own network is called.
func New(
	appRepo repository.AppRepo,
	projectEnvRepo repository.ProjectEnvRepo,
	projectRepo repository.ProjectRepo,
	settingRepo repository.SettingRepo,
	userRepo repository.UserRepo,

	projectService projectservice.Service,
	appService appservice.Service,
	appDeploymentService appdeploymentservice.Service,
	appProvisionService appprovisionservice.Service,
	appRoutingService approutingservice.Service,

	clusterService clusterservice.Service,
	clusterSecretService clustersecretservice.Service,
	domainService domainservice.Service,
	envVarService envvarservice.Service,
	networkService networkservice.Service,
	sslService sslservice.Service,
	volumeService volumeservice.Service,

	taskQueue queue.TaskQueue,
) specservice.Service {
	svc := &service{
		appRepo:        appRepo,
		projectEnvRepo: projectEnvRepo,
		projectRepo:    projectRepo,
		settingRepo:    settingRepo,
		userRepo:       userRepo,

		projectService:       projectService,
		appService:           appService,
		appDeploymentService: appDeploymentService,
		appProvisionService:  appProvisionService,
		appRoutingService:    appRoutingService,

		clusterService:       clusterService,
		clusterSecretService: clusterSecretService,
		domainService:        domainService,
		envVarService:        envVarService,
		networkService:       networkService,
		sslService:           sslService,
		volumeService:        volumeService,

		taskQueue: taskQueue,
	}
	svc.loadOwned = svc.loadOwnedFromRepo
	svc.loadByIDs = svc.loadByIDsFromRepo
	svc.findRef = svc.findRefInRepo
	svc.nodeExists = svc.nodeExistsInRepo
	return svc
}

// settingLoader loads the settings one scope defines, and nothing it merely
// inherits.
//
// It is a named seam rather than a bare repository call for two reasons. The
// ownership rule - a scope exports what it defines, because an outer scope's
// settings belong to the outer scope's document - lives in one place instead of
// being restated at five call sites. And bunex options are opaque closures, so
// a test double given the repository interface cannot tell one scope's query
// from another's; with typed arguments it can.
type settingLoader func(
	ctx context.Context,
	db database.IDB,
	scopes []base.ObjectScopeType,
	objectID string,
) ([]*entity.Setting, error)

// settingsByIDLoader loads settings by id, whatever their scope. Export asks it
// for the settings a reference reaches outside what is exported. It is a seam
// for the reason settingLoader is: a test double cannot read bunex options.
type settingsByIDLoader func(ctx context.Context, db database.IDB, ids []string) ([]*entity.Setting, error)

// refFinder finds the setting an external reference names, as a scope sees
// settings: by id, then by type, name and kind. It answers nil when nothing
// matches. It is a seam for the reason settingLoader is.
type refFinder func(
	ctx context.Context,
	db database.IDB,
	scope *entity.ObjectScope,
	ref *specmodel.ExternalRef,
) (*entity.Setting, error)

// nodeFinder reports whether this installation has a cluster node, by its
// Docker id. It is a seam for the reason settingLoader is.
type nodeFinder func(ctx context.Context, db database.IDB, nodeID string) (bool, error)

type service struct {
	appRepo        repository.AppRepo
	projectEnvRepo repository.ProjectEnvRepo
	projectRepo    repository.ProjectRepo
	settingRepo    repository.SettingRepo
	userRepo       repository.UserRepo

	projectService       projectservice.Service
	appService           appservice.Service
	appDeploymentService appdeploymentservice.Service
	appProvisionService  appprovisionservice.Service
	appRoutingService    approutingservice.Service
	clusterSecretService clustersecretservice.Service
	envVarService        envvarservice.Service
	networkService       networkservice.Service
	taskQueue            queue.TaskQueue

	clusterService clusterservice.Service
	domainService  domainservice.Service
	sslService     sslservice.Service
	volumeService  volumeservice.Service

	loadOwned  settingLoader
	loadByIDs  settingsByIDLoader
	findRef    refFinder
	nodeExists nodeFinder
}
