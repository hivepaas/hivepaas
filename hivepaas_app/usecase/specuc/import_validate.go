package specuc

import (
	"context"
	"fmt"

	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/basedto"
	"github.com/hivepaas/hivepaas/hivepaas_app/config"
	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/permission"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/specservice"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/specservice/specmodel"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/specuc/specdto"
)

// ValidateImport answers what importing a bundle at a scope would do, writing
// nothing.
//
// A bundle that carries secrets passes the gate an export producing one
// passes. Planning it compares this installation's secrets with the bundle's,
// and "same" or "different" is enough to test a guess - so the gate is asked as
// soon as the bundle says what it carries, before anything is compared.
func (uc *UC) ValidateImport(
	ctx context.Context,
	auth *basedto.Auth,
	req *specdto.ValidateImportReq,
) (*specdto.ValidateImportResp, error) {
	resp, err := uc.specService.ValidateImport(ctx, uc.db, uc.importReq(auth, req))
	if err != nil {
		return nil, hperrors.Wrap(err)
	}
	return &specdto.ValidateImportResp{Data: resp.Plan}, nil
}

// importReq is what validate and apply both ask the service, with the gates of
// the caller: revealing secrets, granting capabilities, reaching another app's
// storage, changing a project's owner - and the operator's switch over mounts of
// the host.
func (uc *UC) importReq(auth *basedto.Auth, req *specdto.ValidateImportReq) *specservice.ValidateImportReq {
	cfg := config.Current()
	return &specservice.ValidateImportReq{
		Scope:               req.Scope,
		Bundle:              req.Bundle,
		Passphrase:          req.Passphrase,
		Selection:           req.Selection,
		Options:             req.Options,
		AllowPrivilegedApps: cfg != nil && cfg.Security.AllowPrivilegedApps,
		AuthorizeSecrets: func(ctx context.Context, mode specmodel.SecretsMode) error {
			return uc.permissionManager.AuthorizeSecretReveal(ctx, uc.db, auth, &permission.RevealSubject{
				Scope:    req.Scope.ScopeType,
				ObjectID: req.Scope.ScopeObjectID(),
				Source:   base.AuditLogSourceAPIAction,
				ResType:  base.ResourceTypeSetting,
				ResName:  fmt.Sprintf("configuration spec import (%s)", mode),
			})
		},
		// The gates template creation applies to the same grants: the resources
		// screen's for capabilities, the storage screen's for another app's files.
		MayWriteCluster: func(ctx context.Context) (bool, error) {
			return uc.permissionManager.CheckAccess(ctx, uc.db, auth, &permission.ModuleAccessCheck{
				BaseAccessCheck: permission.BaseAccessCheck{Action: base.ActionTypeWrite},
				Module:          base.ResourceModuleCluster,
			})
		},
		MayWriteApp: func(ctx context.Context, app *entity.App) (bool, error) {
			return uc.permissionManager.CheckAccess(ctx, uc.db, auth, &permission.AppAccessCheck{
				BaseAccessCheck: permission.BaseAccessCheck{Action: base.ActionTypeWrite},
				AppID:           app.ID,
				ParentID:        app.ParentID,
				ProjectID:       app.ProjectID,
				ProjectEnv:      app.ProjectEnvID,
			})
		},
		// Project update's gate: an admin, the current owner, or Write on the
		// Project module.
		MayChangeOwner: func(ctx context.Context, project *entity.Project) (bool, error) {
			if auth.User.IsAdmin() || auth.User.ID == project.OwnerID {
				return true, nil
			}
			return uc.permissionManager.CheckAccess(ctx, uc.db, auth, &permission.ModuleAccessCheck{
				BaseAccessCheck: permission.BaseAccessCheck{Action: base.ActionTypeWrite},
				Module:          base.ResourceModuleProject,
			})
		},
	}
}
