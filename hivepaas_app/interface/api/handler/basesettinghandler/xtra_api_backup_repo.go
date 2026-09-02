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

	resType := base.ResourceTypeBackupRepo
	scope := &entity.ObjectScope{ScopeType: scopeType}
	switch scopeType {
	case base.ObjectScopeProject:
		auth, scope.ProjectID, _, err = h.GetAuthProjectSettings(ctx, base.ActionTypeWrite, "")
	case base.ObjectScopeProjectEnv:
		auth, scope.ProjectID, scope.ProjectEnvID, _, err = h.GetAuthProjectEnvSettings(ctx,
			base.ActionTypeWrite, "")
	case base.ObjectScopeApp:
		auth, scope.ProjectID, scope.ProjectEnvID, scope.AppID, _, err = h.GetAuthAppSettings(ctx,
			base.ActionTypeWrite, "")
	case base.ObjectScopeGlobal, base.ObjectScopeHivepaas:
		auth, _, err = h.GetAuthGlobalSettings(ctx, resType, base.ActionTypeWrite, "")
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
