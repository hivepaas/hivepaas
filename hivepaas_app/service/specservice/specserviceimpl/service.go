package specserviceimpl

import (
	"context"

	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/infra/database"
	"github.com/hivepaas/hivepaas/hivepaas_app/repository"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/clusterservice"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/domainservice"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/specservice"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/specservice/specmodel"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/volumeservice"
)

// New builds the spec exporter.
//
// There is no networkservice dependency on purpose. Resolving a Docker network
// id to its name needs only the cluster-network settings, which sync writes
// with RefID set to the Docker id and Name to the Docker name - so the mapping
// comes from the database rather than from a round trip to Docker.
func New(
	appRepo repository.AppRepo,
	projectEnvRepo repository.ProjectEnvRepo,
	projectRepo repository.ProjectRepo,
	settingRepo repository.SettingRepo,

	clusterService clusterservice.Service,
	domainService domainservice.Service,
	volumeService volumeservice.Service,
) specservice.Service {
	svc := &service{
		appRepo:        appRepo,
		projectEnvRepo: projectEnvRepo,
		projectRepo:    projectRepo,
		settingRepo:    settingRepo,

		clusterService: clusterService,
		domainService:  domainService,
		volumeService:  volumeService,
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

	clusterService clusterservice.Service
	domainService  domainservice.Service
	volumeService  volumeservice.Service

	loadOwned  settingLoader
	loadByIDs  settingsByIDLoader
	findRef    refFinder
	nodeExists nodeFinder
}
