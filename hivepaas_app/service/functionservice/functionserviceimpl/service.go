package functionserviceimpl

import (
	"context"
	"errors"

	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/infra/database"
	"github.com/hivepaas/hivepaas/hivepaas_app/interface/agent/client/functionservice"
	"github.com/hivepaas/hivepaas/hivepaas_app/repository"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/agentservice"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/clusterservice"
	fnservice "github.com/hivepaas/hivepaas/hivepaas_app/service/functionservice"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/functionservice/functiontest"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/imagebuildservice"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/networkservice"
	"github.com/hivepaas/hivepaas/services/docker"
)

// runner runs a test run on this node.
type runner interface {
	Run(ctx context.Context, req *functiontest.RunReq) (*functiontest.RunResp, error)
}

type service struct {
	imageBuildService imagebuildservice.Service
	clusterService    clusterservice.Service
	networkService    networkservice.Service
	agentService      agentservice.Service

	localRunner       runner
	newAgentClient    func(addr string) (functionservice.FunctionServiceClient, error)
	loadBuildSettings func(ctx context.Context, db database.IDB, app *entity.App) (*entity.ImageBuildSettings, error)
}

func New(
	dockerManager docker.Manager,
	settingRepo repository.SettingRepo,
	imageBuildService imagebuildservice.Service,
	clusterService clusterservice.Service,
	networkService networkservice.Service,
	agentService agentservice.Service,
) fnservice.Service {
	return &service{
		imageBuildService: imageBuildService,
		clusterService:    clusterService,
		networkService:    networkService,
		agentService:      agentService,
		localRunner: &functiontest.Runner{
			Docker:  dockerManager,
			Builder: functiontest.NewLibrariesBuilder(imageBuildService),
			TempDir: base.BaseTempDirDefault,
		},
		newAgentClient: functionservice.NewFunctionServiceClient,
		loadBuildSettings: func(ctx context.Context, db database.IDB, app *entity.App) (
			*entity.ImageBuildSettings, error) {
			setting, err := settingRepo.GetSingle(ctx, db, app.Project.GetObjectScope(), base.SettingTypeImageBuild,
				true)
			if err != nil {
				if errors.Is(err, hperrors.ErrNotFound) {
					return nil, nil
				}
				return nil, hperrors.Wrap(err)
			}
			return setting.MustAsImageBuildSettings(), nil
		},
	}
}
