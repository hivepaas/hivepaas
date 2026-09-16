package specserviceimpl

import (
	"context"

	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/infra/database"
	"github.com/hivepaas/hivepaas/hivepaas_app/repository"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/clusterservice"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/specservice"
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
) specservice.Service {
	svc := &service{
		appRepo:        appRepo,
		projectEnvRepo: projectEnvRepo,
		projectRepo:    projectRepo,
		settingRepo:    settingRepo,

		clusterService: clusterService,
	}
	svc.loadOwned = svc.loadOwnedFromRepo
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

type service struct {
	appRepo        repository.AppRepo
	projectEnvRepo repository.ProjectEnvRepo
	projectRepo    repository.ProjectRepo
	settingRepo    repository.SettingRepo

	clusterService clusterservice.Service

	loadOwned settingLoader
}
