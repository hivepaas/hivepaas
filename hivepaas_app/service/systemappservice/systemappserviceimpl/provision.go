package systemappserviceimpl

import (
	"context"

	"github.com/moby/moby/api/types/swarm"
	"github.com/tiendc/gofn"

	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/infra/database"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/bunex"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/timeutil"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/ulid"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/appprovisionservice"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/specservice"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/systemappservice"
)

func (s *service) Provision(
	ctx context.Context,
	db database.IDB,
	req *systemappservice.ProvisionReq,
) (*systemappservice.ProvisionResp, error) {
	project, projectEnv, err := s.projectEnv(ctx, db, gofn.Coalesce(req.Env, systemappservice.DefaultEnv))
	if err != nil {
		return nil, hperrors.Wrap(err)
	}

	timeNow := timeutil.NowUTC()
	provisioned, err := s.provisionService.ProvisionApps(ctx, db, &appprovisionservice.ProvisionAppsReq{
		Apps: []*appprovisionservice.ProvisionAppReq{{
			ProjectID:    project.ID,
			ProjectEnvID: projectEnv.ID,
			AppID:        gofn.Must(ulid.NewStringULID()),
			Key:          req.Key,
			Name:         req.Name,
			Status:       base.AppStatusActive,
			Note:         req.Note,
			Configure: func(ctx context.Context, db database.IDB, app *entity.App,
				spec *swarm.ServiceSpec) ([]*entity.Setting, error) {
				built, buildErr := s.specService.BuildApp(ctx, db, &specservice.BuildAppReq{
					App: app, Doc: req.Doc, Spec: spec, TimeNow: timeNow,
				})
				if buildErr != nil {
					return nil, hperrors.Wrap(buildErr)
				}
				if req.Customize != nil {
					if err := req.Customize(spec); err != nil {
						return nil, hperrors.Wrap(err)
					}
				}
				return built.Settings, nil
			},
			Deployment: &appprovisionservice.FirstDeployment{
				Source:   base.DeploymentTriggerSourceAPI,
				SourceID: req.TriggerUserID,
			},
		}},
	})
	resp := &systemappservice.ProvisionResp{}
	if provisioned != nil {
		resp.Cleanup = provisioned.Cleanup
	}
	if err != nil {
		return resp, hperrors.Wrap(err)
	}

	one := provisioned.Apps[0]
	resp.App, resp.DeploymentTask, resp.CertTasks = one.App, one.DeploymentTask, one.CertTasks
	return resp, nil
}

// projectEnv returns the hidden hivepaas project and one of its environments,
// creating the environment the first time a system app asks for it.
//
// The project itself is created at startup, with the stack's own services in it,
// so a missing project is a broken installation rather than something to repair
// here.
func (s *service) projectEnv(ctx context.Context, db database.IDB, env string) (
	*entity.Project, *entity.ProjectEnv, error) {
	project, err := s.projectRepo.GetByKey(ctx, db, base.HivepaasProjectKey,
		bunex.SelectRelation("ProjectEnvs"),
	)
	if err != nil {
		return nil, nil, hperrors.Wrap(err)
	}
	if projectEnv := project.GetEnv(env); projectEnv != nil {
		return project, projectEnv, nil
	}

	projectEnv := project.GetOrCreateEnv(env)
	err = s.projectEnvRepo.Upsert(ctx, db, projectEnv,
		entity.ProjectEnvUpsertingConflictCols, entity.ProjectEnvUpsertingUpdateCols)
	if err != nil {
		return nil, nil, hperrors.Wrap(err)
	}
	return project, projectEnv, nil
}
