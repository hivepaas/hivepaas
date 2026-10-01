package appuc

import (
	"context"
	"errors"
	"time"

	"github.com/moby/moby/api/types/swarm"
	"github.com/tiendc/gofn"

	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/basedto"
	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/infra/database"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/auditdetail"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/transaction"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/ulid"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/appdeploymentservice"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/appprovisionservice"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/appuc/appdto"
)

// CreateFunction creates a function: an app of kind function whose deployment
// settings are its source, deployed as soon as it exists. An app becomes a
// function only this way.
func (uc *UC) CreateFunction(
	ctx context.Context,
	auth *basedto.Auth,
	req *appdto.CreateFunctionReq,
) (resp *appdto.CreateFunctionResp, err error) {
	source, err := req.Source.ToEntity()
	if err != nil {
		return nil, hperrors.Wrap(err)
	}
	// Before the transaction: checking the repository reaches it, and nothing is
	// locked while it does.
	if err = uc.checkFunctionSource(ctx, req.ProjectID, req.ProjectEnvID, source); err != nil {
		return nil, hperrors.Wrap(err)
	}

	var provisioned *appprovisionservice.ProvisionAppsResp
	committed := false
	defer func() {
		if rec := recover(); rec != nil {
			err = errors.Join(err, hperrors.NewPanic(rec))
		}
		// The swarm service is created inside the transaction but is not part of
		// it: a rolled-back request would leave it behind with nothing recorded.
		if err != nil && !committed && provisioned != nil {
			if cleanupErr := provisioned.Cleanup(context.WithoutCancel(ctx)); cleanupErr != nil {
				err = errors.Join(err, cleanupErr)
			}
		}
	}()

	var created *appprovisionservice.ProvisionAppResp
	err = transaction.Execute(ctx, uc.db, func(db database.Tx) error {
		var txErr error
		provisioned, txErr = uc.appProvisionService.ProvisionApps(ctx, db, &appprovisionservice.ProvisionAppsReq{
			Apps: []*appprovisionservice.ProvisionAppReq{{
				ProjectID:    req.ProjectID,
				ProjectEnvID: req.ProjectEnvID,
				Name:         req.Name,
				Status:       req.Status,
				Note:         req.Note,
				Tags:         req.Tags,
				Configure: func(_ context.Context, _ database.IDB, app *entity.App, _ *swarm.ServiceSpec) (
					[]*entity.Setting, error) {
					return functionSettings(app, source, app.CreatedAt), nil
				},
				Deployment: &appprovisionservice.FirstDeployment{
					Source:   base.DeploymentTriggerSourceUser,
					SourceID: auth.User.ID,
				},
			}},
		})
		if txErr != nil {
			return hperrors.Wrap(txErr)
		}
		created = provisioned.Apps[0]
		return uc.recordAppWrite(ctx, db, auth, created.App,
			base.AuditLogTypeAppCreate, base.AuditLogSourceAPICreate, "create", auditdetail.New().
				Set("projectId", created.App.ProjectID).
				Set("envId", created.App.ProjectEnvID).
				Set("serviceId", created.App.ServiceID).
				Set("runtime", string(source.Runtime)))
	})
	if err != nil {
		return nil, hperrors.Wrap(err)
	}
	committed = true

	// A task can be picked up only once its row exists, which is once the
	// transaction has committed.
	if err = uc.taskQueue.ScheduleTask(ctx, created.DeploymentTask); err != nil {
		return nil, hperrors.Wrap(err)
	}
	if err = uc.taskQueue.ScheduleTask(ctx, created.CertTasks...); err != nil {
		return nil, hperrors.Wrap(err)
	}

	return &appdto.CreateFunctionResp{Data: &appdto.CreateFunctionDataResp{
		ID:           created.App.ID,
		DeploymentID: created.Deployment.ID,
		TaskID:       created.DeploymentTask.ID,
	}}, nil
}

// checkFunctionSource checks a function's source in the environment it is
// created in: the settings it refers to exist there and are active, and its
// build has what it will need.
func (uc *UC) checkFunctionSource(
	ctx context.Context,
	projectID, projectEnvID string,
	source *entity.DeploymentFunctionSource,
) error {
	deployment := &entity.AppDeploymentSettings{ActiveMethod: base.DeploymentMethodFunction, FunctionSource: source}
	env := &entity.ProjectEnv{ID: projectEnvID, ProjectID: projectID}
	refObjects := entity.NewRefObjects()
	err := uc.settingService.LoadRefObjectsByIDs(ctx, uc.db, &refObjects, env.GetObjectScope(), true,
		deployment.GetRefObjectIDs())
	if err != nil {
		return hperrors.Wrap(err)
	}

	checkReq := &appdeploymentservice.CheckBuildSourceReq{PushToRegistry: source.PushToRegistry, RefObjects: refObjects}
	if source.Code.Repo != nil {
		checkReq.RepoSource = source.Code.Repo.RepoSource()
	}
	return hperrors.Wrap(uc.appDeploymentService.CheckBuildSource(ctx, checkReq))
}

// functionSettings are what a function is created with: its kind, its source
// to deploy, and its routing at the port its runtime listens on.
func functionSettings(app *entity.App, source *entity.DeploymentFunctionSource, timeNow time.Time) []*entity.Setting {
	newSetting := func(typ base.SettingType, version int, inheritable bool, data entity.SettingData) *entity.Setting {
		setting := &entity.Setting{
			ID:          gofn.Must(ulid.NewStringULID()),
			Scope:       base.ObjectScopeApp,
			ObjectID:    app.ID,
			Type:        typ,
			Status:      base.SettingStatusActive,
			Inheritable: inheritable,
			Version:     version,
			CreatedAt:   timeNow,
			UpdatedAt:   timeNow,
		}
		setting.MustSetData(data)
		return setting
	}
	return []*entity.Setting{
		newSetting(base.SettingTypeAppKind, entity.CurrentAppKindSettingsVersion, false,
			&entity.AppKindSettings{Category: base.AppCategoryFunction}),
		newSetting(base.SettingTypeAppDeployment, entity.CurrentAppDeploymentSettingsVersion, true,
			&entity.AppDeploymentSettings{ActiveMethod: base.DeploymentMethodFunction, FunctionSource: source}),
		newSetting(base.SettingTypeAppRouting, entity.CurrentAppRoutingSettingsVersion, true,
			&entity.AppRoutingSettings{Port: base.FunctionPort}),
	}
}
