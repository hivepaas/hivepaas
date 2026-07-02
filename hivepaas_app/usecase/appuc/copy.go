package appuc

import (
	"context"

	"github.com/moby/moby/api/types/swarm"
	"github.com/tiendc/gofn"

	"github.com/hivepaas/hivepaas/hivepaas_app/apperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/basedto"
	"github.com/hivepaas/hivepaas/hivepaas_app/config"
	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/infra/database"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/bunex"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/transaction"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/appcopyservice"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/appuc/appdto"
)

func (uc *UC) CopyApp(
	ctx context.Context,
	auth *basedto.Auth,
	req *appdto.CopyAppReq,
) (*appdto.CopyAppResp, error) {
	var data *copyAppData
	var copyResp *appcopyservice.AppCopyResp

	err := transaction.Execute(ctx, uc.db, func(db database.Tx) error {
		data = &copyAppData{}
		err := uc.loadAppDataForCopying(ctx, db, req, data)
		if err != nil {
			return apperrors.New(err)
		}

		copyResp, err = uc.appCopyService.CopyApp(ctx, db, &appcopyservice.AppCopyReq{
			SrcProject:    data.Project,
			SrcApp:        data.App,
			TargetProject: data.Project,
			OnCopyApp: func(targetApp, srcApp *entity.App) error {
				return uc.onCopyApp(req, targetApp, srcApp)
			},
			OnCopySetting: func(targetApp *entity.App, setting *entity.Setting) (*entity.Setting, error) {
				return uc.onCopyAppSetting(req, setting)
			},
			OnCopyService: func(targetSvc, srcSvc *swarm.Service) error {
				return uc.onCopyAppService(req, targetSvc, srcSvc)
			},
		})
		if err != nil {
			return apperrors.New(err)
		}
		return nil
	})
	// Run the cleanup function
	if copyResp != nil && copyResp.OnCleanup != nil {
		_ = copyResp.OnCleanup(err)
	}
	if err != nil {
		return nil, apperrors.New(err)
	}

	return &appdto.CopyAppResp{
		Data: &basedto.ObjectIDResp{ID: copyResp.TargetApp.ID},
	}, nil
}

type copyAppData struct {
	Project *entity.Project
	App     *entity.App
}

func (uc *UC) loadAppDataForCopying(
	ctx context.Context,
	db database.IDB,
	req *appdto.CopyAppReq,
	data *copyAppData,
) error {
	app, err := uc.appService.LoadApp(ctx, db, req.ProjectID, req.AppID, false, false,
		bunex.SelectFor("UPDATE OF app"),
		bunex.SelectRelation("Project"),
	)
	if err != nil {
		return apperrors.New(err)
	}
	if app.UpdateVer != req.UpdateVer {
		return apperrors.New(apperrors.ErrUpdateVerMismatched)
	}

	data.App = app
	data.Project = app.Project
	return nil
}

func (uc *UC) onCopyApp(
	req *appdto.CopyAppReq,
	targetApp, _ *entity.App,
) error {
	targetApp.Name = req.TargetName
	targetApp.Env = req.TargetEnv
	targetApp.Status = gofn.Coalesce(req.TargetStatus, base.AppStatusActive)
	return nil
}

func (uc *UC) onCopyAppSetting(
	req *appdto.CopyAppReq,
	setting *entity.Setting,
) (*entity.Setting, error) {
	switch setting.Type { //nolint:exhaustive
	case base.SettingTypeApp:
		return setting, nil
	case base.SettingTypeAppDeployment:
		return uc.onCopyDeploymentSetting(req, setting)
	case base.SettingTypeAppFeatures:
		return setting, nil
	case base.SettingTypeAppHttp:
		return uc.onCopyHttpSetting(req, setting)
	case base.SettingTypeConfigFile:
		return gofn.If(req.CopyConfigFiles.Copy, setting, nil), nil
	case base.SettingTypeEnvVar:
		return gofn.If(req.CopyEnvVars.Copy, setting, nil), nil
	case base.SettingTypeHealthcheck:
		return gofn.If(req.CopyHealthChecks.Copy, setting, nil), nil
	case base.SettingTypeSchedJob:
		return gofn.If(req.CopySchedJobs.Copy, setting, nil), nil
	case base.SettingTypeSecret:
		return gofn.If(req.CopySecrets.Copy, setting, nil), nil
	default:
		return nil, nil
	}
}

func (uc *UC) onCopyDeploymentSetting(
	req *appdto.CopyAppReq,
	setting *entity.Setting,
) (*entity.Setting, error) {
	deploymentSettings := setting.MustAsAppDeploymentSettings()

	if !req.CopyDeploymentSettings.Copy {
		isDevEnv := config.Current.IsDevEnv()
		deploymentSettings.ActiveMethod = base.DeploymentMethodImage
		deploymentSettings.ImageSource = &entity.DeploymentImageSource{
			Image: gofn.If(isDevEnv, dockerImageInitDev, dockerImageInit),
		}
		deploymentSettings.Command = gofn.If(isDevEnv, "sleep infinity", "")
		deploymentSettings.WorkingDir = ""
		deploymentSettings.PreDeploymentCommand = ""
		deploymentSettings.PostDeploymentCommand = ""
	}

	setting.MustSetData(deploymentSettings)
	return setting, nil
}

func (uc *UC) onCopyHttpSetting(
	req *appdto.CopyAppReq,
	setting *entity.Setting,
) (*entity.Setting, error) {
	httpSettings := setting.MustAsAppHttpSettings()

	currDomains := httpSettings.Domains
	httpSettings.Domains = nil
	for _, copySettings := range req.CopyHttpSettings.CopyDomainSettings {
		appDomain, _ := gofn.Find(currDomains, func(item *entity.AppDomain) bool {
			return item.Domain == copySettings.SourceDomain
		})
		if appDomain == nil {
			continue
		}
		appDomain.Domain = copySettings.TargetDomain
		appDomain.SSLCert = entity.ObjectID{ID: copySettings.TargetSSLCert.ID}
		// TODO: handle SSL cert validation
		httpSettings.Domains = append(httpSettings.Domains, appDomain)
	}

	setting.MustSetData(httpSettings)
	return setting, nil
}

func (uc *UC) onCopyAppService(
	req *appdto.CopyAppReq,
	targetSvc, _ *swarm.Service,
) error {
	targetSvcSpec := &targetSvc.Spec
	containerSpec := targetSvcSpec.TaskTemplate.ContainerSpec
	isDevEnv := config.Current.IsDevEnv()
	if !req.CopyDeploymentSettings.Copy {
		containerSpec.Image = gofn.If(isDevEnv, dockerImageInitDev, dockerImageInit)
		containerSpec.Command = gofn.If(isDevEnv, nil, []string{"sleep", "infinity"})
		containerSpec.Args = nil
		containerSpec.Dir = ""
	}

	return nil
}
