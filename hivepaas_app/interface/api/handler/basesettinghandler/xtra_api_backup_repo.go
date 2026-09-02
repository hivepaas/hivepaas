package basesettinghandler

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/basedto"
	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/settings/backuprepouc/backuprepodto"
)

func (h *Handler) BackupRepoChangePassword(ctx *gin.Context, scopeType base.ObjectScopeType) {
	var auth *basedto.Auth
	var err error
	var itemID string

	resType := base.ResourceTypeBackupRepo
	scope := &entity.ObjectScope{ScopeType: scopeType}
	switch scopeType {
	case base.ObjectScopeProject:
		auth, scope.ProjectID, itemID, err = h.GetAuthProjectSettings(ctx, base.ActionTypeWrite, "itemID")
	case base.ObjectScopeProjectEnv:
		auth, scope.ProjectID, scope.ProjectEnvID, itemID, err = h.GetAuthProjectEnvSettings(ctx,
			base.ActionTypeWrite, "itemID")
	case base.ObjectScopeApp:
		auth, scope.ProjectID, scope.ProjectEnvID, scope.AppID, itemID, err = h.GetAuthAppSettings(ctx,
			base.ActionTypeWrite, "itemID")
	case base.ObjectScopeGlobal, base.ObjectScopeHivepaas:
		auth, itemID, err = h.GetAuthGlobalSettings(ctx, resType, base.ActionTypeWrite, "itemID")
	case base.ObjectScopeUser:
		fallthrough
	default:
		h.RenderError(ctx, hperrors.Wrap(hperrors.ErrObjectScopeInvalid).WithParam("Scope", scopeType))
		return
	}
	if err != nil {
		h.RenderError(ctx, err)
		return
	}

	req := backuprepodto.NewChangeRepoPasswordReq()
	req.Scope = scope
	req.Type = base.SettingTypeBackupRepo
	req.ID = itemID
	if err = h.ParseAndValidateJSONBody(ctx, req); err != nil {
		h.RenderError(ctx, err)
		return
	}

	resp, err := h.BackupRepoUC.ChangeRepoPassword(h.RequestCtx(ctx), auth, req)
	if err != nil {
		h.RenderError(ctx, err)
		return
	}

	ctx.JSON(http.StatusOK, resp)
}

func (h *Handler) BackupRepoCleanup(ctx *gin.Context, scopeType base.ObjectScopeType) {
	var auth *basedto.Auth
	var itemID string
	var err error

	resType := base.ResourceTypeBackupRepo
	scope := &entity.ObjectScope{ScopeType: scopeType}
	switch scopeType {
	case base.ObjectScopeProject:
		auth, scope.ProjectID, itemID, err = h.GetAuthProjectSettings(ctx, base.ActionTypeWrite, "itemID")
	case base.ObjectScopeProjectEnv:
		auth, scope.ProjectID, scope.ProjectEnvID, itemID, err = h.GetAuthProjectEnvSettings(ctx,
			base.ActionTypeWrite, "itemID")
	case base.ObjectScopeApp:
		auth, scope.ProjectID, scope.ProjectEnvID, scope.AppID, itemID, err = h.GetAuthAppSettings(ctx,
			base.ActionTypeWrite, "itemID")
	case base.ObjectScopeGlobal, base.ObjectScopeHivepaas:
		auth, itemID, err = h.GetAuthGlobalSettings(ctx, resType, base.ActionTypeWrite, "itemID")
	case base.ObjectScopeUser:
		fallthrough
	default:
		h.RenderError(ctx, hperrors.Wrap(hperrors.ErrObjectScopeInvalid).WithParam("Scope", scopeType))
		return
	}
	if err != nil {
		h.RenderError(ctx, err)
		return
	}

	req := backuprepodto.NewCleanupBackupRepoReq()
	req.Scope = scope
	req.Type = base.SettingTypeBackupRepo
	req.ID = itemID
	// The request carries no body, so validate it here rather than through the body parser.
	if vldErrs := req.Validate(); len(vldErrs) > 0 {
		h.RenderError(ctx, vldErrs)
		return
	}

	resp, err := h.BackupRepoUC.CleanupBackupRepo(h.RequestCtx(ctx), auth, req)
	if err != nil {
		h.RenderError(ctx, err)
		return
	}

	ctx.JSON(http.StatusOK, resp)
}
