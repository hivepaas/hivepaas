package appsettingsuc

import (
	"context"

	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/basedto"
	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/permission"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/auditdetail"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/bunex"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/appsettingsuc/appsettingsdto"
)

func (uc *UC) GetAppContainerSettings(
	ctx context.Context,
	auth *basedto.Auth,
	req *appsettingsdto.GetAppContainerSettingsReq,
) (*appsettingsdto.GetAppContainerSettingsResp, error) {
	app, err := uc.appService.LoadApp(ctx, uc.db, req.ProjectID, req.AppID, false, false,
		bunex.SelectExcludeColumns(entity.AppDefaultExcludeColumns...),
		bunex.SelectRelation("Project",
			bunex.SelectExcludeColumns(entity.ProjectDefaultExcludeColumns...),
		),
		bunex.SelectRelation("ProjectEnv"),
	)
	if err != nil {
		return nil, hperrors.Wrap(err)
	}

	// Gated before the service is read, so that a refusal reads nothing. System
	// labels can carry what traefik is told - a basic auth user list among it -
	// so they take what secrets take, and the attempt is recorded either way.
	if req.RevealSystemLabels {
		err = uc.permissionManager.AuthorizeSecretReveal(ctx, uc.db, auth, &permission.RevealSubject{
			Scope:    base.ObjectScopeApp,
			ObjectID: app.ID,
			Source:   base.AuditLogSourceAPIGet,
			ResType:  base.ResourceTypeApp,
			ResID:    app.ID,
			ResName:  app.Name,
			Detail:   auditdetail.New().Set("revealed", "systemLabels").String(),
		})
		if err != nil {
			return nil, hperrors.Wrap(err)
		}
	}

	service, err := uc.clusterService.ServiceInspect(ctx, app.ServiceID, true)
	if err != nil {
		return nil, hperrors.Wrap(err)
	}
	if service, err = requireAppService(service, app.ID); err != nil {
		return nil, err
	}

	resp, err := appsettingsdto.TransformContainerSettings(service, req.RevealSystemLabels)
	if err != nil {
		return nil, hperrors.Wrap(err)
	}

	return &appsettingsdto.GetAppContainerSettingsResp{
		Data: resp,
	}, nil
}
